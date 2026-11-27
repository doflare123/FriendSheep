package pushdevices

import (
	"context"
	"errors"
)

const maxAdmissionVersionChanges = 4

var ErrAdmissionChanged = errors.New("push-привязка менялась во время допуска")

type admissionStore interface {
	ResolveActiveAddress(ctx context.Context, installationID, bindingID string, userID uint64) (ResolvedAddress, bool, error)
	DetachIfCurrent(ctx context.Context, binding Binding) (bool, error)
}

// AddressAuthorizer допускает будущую внешнюю push-попытку только для текущего
// адреса и действующей пары пользователь/сессия. Сам внешний вызов сюда не входит.
type AddressAuthorizer struct {
	store   admissionStore
	checker SessionChecker
}

func NewAddressAuthorizer(store admissionStore, checker SessionChecker) (*AddressAuthorizer, error) {
	if store == nil || checker == nil {
		return nil, errors.New("допуск push-адреса не настроен")
	}
	return &AddressAuthorizer{store: store, checker: checker}, nil
}

func (a *AddressAuthorizer) Authorize(ctx context.Context, installationID, bindingID string, userID uint64) (ResolvedAddress, bool, error) {
	for attempt := 0; attempt < maxAdmissionVersionChanges; attempt++ {
		address, found, err := a.store.ResolveActiveAddress(ctx, installationID, bindingID, userID)
		if err != nil || !found {
			return ResolvedAddress{}, false, err
		}
		active, err := a.checker.CheckSession(ctx, address.UserID, address.SessionID)
		if err != nil {
			return ResolvedAddress{}, false, err
		}
		if !active {
			if _, err := a.store.DetachIfCurrent(ctx, address.Binding); err != nil {
				return ResolvedAddress{}, false, err
			}
			return ResolvedAddress{}, false, nil
		}

		current, stillActive, err := a.store.ResolveActiveAddress(ctx, installationID, bindingID, userID)
		if err != nil || !stillActive {
			return ResolvedAddress{}, false, err
		}
		if current.Version == address.Version && current.SessionID == address.SessionID {
			if err := ctx.Err(); err != nil {
				return ResolvedAddress{}, false, err
			}
			return current, true, nil
		}
		// Ротация адреса в том же поколении требует новой проверки сессии.
	}
	return ResolvedAddress{}, false, ErrAdmissionChanged
}
