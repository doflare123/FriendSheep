package pushdevices

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

const (
	DefaultReconciliationInterval = 15 * time.Minute
	MaxReconciliationPageSize     = 100
)

type reconciliationStore interface {
	ListBindingsForReconciliation(ctx context.Context, afterInstallationID string, limit int) ([]Binding, error)
	DetachIfCurrent(ctx context.Context, binding Binding) (bool, error)
}

type ReconciliationResult struct {
	Checked        int
	Detached       int
	TemporaryError int
}

type Reconciler struct {
	logger   *slog.Logger
	store    reconciliationStore
	checker  SessionChecker
	interval time.Duration
	pageSize int
}

func NewReconciler(logger *slog.Logger, store reconciliationStore, checker SessionChecker, interval time.Duration, pageSize int) (*Reconciler, error) {
	if store == nil || checker == nil {
		return nil, errors.New("сверка push-привязок не настроена")
	}
	if interval <= 0 {
		return nil, errors.New("интервал сверки push-привязок должен быть положительным")
	}
	if pageSize < 1 || pageSize > MaxReconciliationPageSize {
		return nil, errors.New("размер страницы сверки push-привязок должен быть от 1 до 100")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Reconciler{logger: logger, store: store, checker: checker, interval: interval, pageSize: pageSize}, nil
}

// Sweep выполняет один полный проход. Ошибка отдельной проверки не прерывает
// страницу: строка останется для следующего полного прохода.
func (r *Reconciler) Sweep(ctx context.Context) (ReconciliationResult, error) {
	var result ReconciliationResult
	after := ""
	for {
		bindings, err := r.store.ListBindingsForReconciliation(ctx, after, r.pageSize)
		if err != nil {
			return result, err
		}
		for _, binding := range bindings {
			result.Checked++
			active, err := r.checker.CheckSession(ctx, binding.UserID, binding.SessionID)
			if err != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				result.TemporaryError++
				continue
			}
			if active {
				continue
			}
			applied, err := r.store.DetachIfCurrent(ctx, binding)
			if err != nil {
				return result, err
			}
			if applied {
				result.Detached++
			}
		}
		if len(bindings) < r.pageSize {
			return result, nil
		}
		after = bindings[len(bindings)-1].InstallationID
	}
}

func (r *Reconciler) Run(ctx context.Context) {
	for {
		result, err := r.Sweep(ctx)
		if err != nil && ctx.Err() == nil {
			r.logger.Warn("сверка push-привязок завершилась ошибкой", "ошибка", err)
		} else if err == nil {
			r.logger.Info("сверка push-привязок завершена",
				"проверено", result.Checked,
				"отвязано", result.Detached,
				"временных_ошибок", result.TemporaryError)
		}
		if ctx.Err() != nil {
			return
		}
		timer := time.NewTimer(r.interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}
	}
}
