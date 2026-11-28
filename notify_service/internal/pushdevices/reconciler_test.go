package pushdevices

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type reconciliationStoreStub struct {
	mu       sync.Mutex
	bindings []Binding
	detached []Binding
}

func (s *reconciliationStoreStub) ListBindingsForReconciliation(_ context.Context, after string, limit int) ([]Binding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	page := make([]Binding, 0, limit)
	for _, binding := range s.bindings {
		if binding.InstallationID > after && len(page) < limit {
			page = append(page, binding)
		}
	}
	return page, nil
}

func (s *reconciliationStoreStub) DetachIfCurrent(_ context.Context, binding Binding) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detached = append(s.detached, binding)
	return true, nil
}

type managerStoreStub struct {
	reconciliationStoreStub
}

func (s *managerStoreStub) ResolveActiveAddress(context.Context, string, string, uint64) (ResolvedAddress, bool, error) {
	return ResolvedAddress{}, false, nil
}

func TestReconcilerSweepsPagesAndSkipsTemporaryErrors(t *testing.T) {
	bindings := []Binding{
		testBinding("install-a", 1),
		testBinding("install-b", 2),
		testBinding("install-c", 3),
	}
	store := &reconciliationStoreStub{bindings: bindings}
	checker := &sessionCheckerStub{results: []sessionCheckStubResult{
		{active: true}, {active: false}, {err: errors.New("временная ошибка")},
	}}
	reconciler, err := NewReconciler(slog.New(slog.NewTextHandler(io.Discard, nil)), store, checker, time.Hour, 2)
	if err != nil {
		t.Fatal(err)
	}
	result, err := reconciler.Sweep(context.Background())
	if err != nil || result != (ReconciliationResult{Checked: 3, Detached: 1, TemporaryError: 1}) {
		t.Fatalf("Sweep() = %#v, %v", result, err)
	}
	if len(store.detached) != 1 || store.detached[0] != bindings[1] {
		t.Fatalf("отвязки = %#v", store.detached)
	}
}

func TestReconcilerDetachesOnlyConfirmedInactiveHTTPSession(t *testing.T) {
	tests := []struct {
		name             string
		response         string
		wantDetached     int
		wantTemporaryErr int
	}{
		{
			name:             "null active is malformed and keeps binding",
			response:         `{"schemaVersion":1,"active":null}`,
			wantTemporaryErr: 1,
		},
		{
			name:         "false active detaches binding",
			response:     `{"schemaVersion":1,"active":false}`,
			wantDetached: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(test.response))
			}))
			defer server.Close()

			checker, err := NewHTTPMonolithSessionChecker(server.URL, "internal-secret", time.Second, nil)
			if err != nil {
				t.Fatal(err)
			}
			binding := testBinding("install-a", 1)
			store := &reconciliationStoreStub{bindings: []Binding{binding}}
			reconciler, err := NewReconciler(slog.New(slog.NewTextHandler(io.Discard, nil)), store, checker, time.Hour, 1)
			if err != nil {
				t.Fatal(err)
			}

			result, err := reconciler.Sweep(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if result != (ReconciliationResult{Checked: 1, Detached: test.wantDetached, TemporaryError: test.wantTemporaryErr}) {
				t.Fatalf("Sweep() = %#v", result)
			}
			if len(store.detached) != test.wantDetached {
				t.Fatalf("число отвязок = %d, ожидалось %d", len(store.detached), test.wantDetached)
			}
			if test.wantDetached == 1 && store.detached[0] != binding {
				t.Fatalf("отвязанная привязка = %#v, ожидалась %#v", store.detached[0], binding)
			}
		})
	}
}

func TestManagerRunsFirstSweepImmediatelyAndStops(t *testing.T) {
	store := &managerStoreStub{reconciliationStoreStub: reconciliationStoreStub{bindings: []Binding{testBinding("install-a", 1)}}}
	checked := make(chan struct{}, 1)
	checker := sessionCheckerFunc(func(context.Context, uint64, string) (bool, error) {
		select {
		case checked <- struct{}{}:
		default:
		}
		return true, nil
	})
	manager, err := NewManager(slog.New(slog.NewTextHandler(io.Discard, nil)), store, checker, time.Hour, MaxReconciliationPageSize)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-checked:
	case <-time.After(time.Second):
		t.Fatal("первый проход не начался сразу после Start")
	}
	if err := manager.Stop(time.Second); err != nil {
		t.Fatal(err)
	}
}

type sessionCheckerFunc func(context.Context, uint64, string) (bool, error)

func (f sessionCheckerFunc) CheckSession(ctx context.Context, userID uint64, sessionID string) (bool, error) {
	return f(ctx, userID, sessionID)
}

func testBinding(installationID string, suffix byte) Binding {
	return Binding{
		InstallationID: installationID,
		BindingID:      testOperationID(uint64(suffix + 10)),
		UserID:         uint64(suffix),
		SessionID:      testSessionID(suffix),
		Version:        testOperationID(uint64(suffix + 20)),
	}
}
