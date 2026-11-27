package pushdevices

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type admissionStoreStub struct {
	mu       sync.Mutex
	resolved []resolvedStubResult
	detached []Binding
}

type resolvedStubResult struct {
	address ResolvedAddress
	found   bool
	err     error
}

func (s *admissionStoreStub) ResolveActiveAddress(context.Context, string, string, uint64) (ResolvedAddress, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.resolved) == 0 {
		return ResolvedAddress{}, false, nil
	}
	result := s.resolved[0]
	s.resolved = s.resolved[1:]
	return result.address, result.found, result.err
}

func (s *admissionStoreStub) DetachIfCurrent(_ context.Context, binding Binding) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detached = append(s.detached, binding)
	return true, nil
}

type sessionCheckerStub struct {
	mu       sync.Mutex
	results  []sessionCheckStubResult
	sessions []string
}

type sessionCheckStubResult struct {
	active bool
	err    error
}

func (s *sessionCheckerStub) CheckSession(_ context.Context, _ uint64, sessionID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = append(s.sessions, sessionID)
	if len(s.results) == 0 {
		return false, errors.New("результат проверки не задан")
	}
	result := s.results[0]
	s.results = s.results[1:]
	return result.active, result.err
}

func TestAddressAuthorizerRepeatsAdmissionAfterAddressRotation(t *testing.T) {
	v1 := testResolvedAddress("d0000000-0000-4000-8000-000000000071", "token-v1")
	v2 := v1
	v2.Version = "d0000000-0000-4000-8000-000000000072"
	v2.ProviderToken = "token-v2"
	store := &admissionStoreStub{resolved: []resolvedStubResult{
		{address: v1, found: true}, {address: v2, found: true},
		{address: v2, found: true}, {address: v2, found: true},
	}}
	checker := &sessionCheckerStub{results: []sessionCheckStubResult{{active: true}, {active: true}}}
	authorizer, err := NewAddressAuthorizer(store, checker)
	if err != nil {
		t.Fatal(err)
	}
	address, allowed, err := authorizer.Authorize(context.Background(), v1.InstallationID, v1.BindingID, v1.UserID)
	if err != nil || !allowed || address.Version != v2.Version || address.ProviderToken != "token-v2" {
		t.Fatalf("Authorize() = %#v, %t, %v", address, allowed, err)
	}
	if len(checker.sessions) != 2 {
		t.Fatalf("проверок сессии = %d, ожидалось 2", len(checker.sessions))
	}
}

func TestAddressAuthorizerDetachesOnlyAfterConfirmedInactive(t *testing.T) {
	address := testResolvedAddress("d0000000-0000-4000-8000-000000000073", "token")
	store := &admissionStoreStub{resolved: []resolvedStubResult{{address: address, found: true}}}
	checker := &sessionCheckerStub{results: []sessionCheckStubResult{{active: false}}}
	authorizer, _ := NewAddressAuthorizer(store, checker)
	resolved, allowed, err := authorizer.Authorize(context.Background(), address.InstallationID, address.BindingID, address.UserID)
	if err != nil || allowed || resolved.ProviderToken != "" || len(store.detached) != 1 || store.detached[0] != address.Binding {
		t.Fatalf("Authorize() = %#v, %t, %v, detached=%#v", resolved, allowed, err, store.detached)
	}
}

func TestAddressAuthorizerDoesNotDetachOnTemporaryCheckError(t *testing.T) {
	address := testResolvedAddress("d0000000-0000-4000-8000-000000000074", "token")
	store := &admissionStoreStub{resolved: []resolvedStubResult{{address: address, found: true}}}
	temporary := errors.New("монолит временно недоступен")
	checker := &sessionCheckerStub{results: []sessionCheckStubResult{{err: temporary}}}
	authorizer, _ := NewAddressAuthorizer(store, checker)
	_, allowed, err := authorizer.Authorize(context.Background(), address.InstallationID, address.BindingID, address.UserID)
	if allowed || !errors.Is(err, temporary) || len(store.detached) != 0 {
		t.Fatalf("Authorize() = allowed:%t error:%v detached=%#v", allowed, err, store.detached)
	}
}

func TestAddressAuthorizerRejectsCancellationAfterFinalRead(t *testing.T) {
	address := testResolvedAddress("d0000000-0000-4000-8000-000000000075", "token")
	ctx, cancel := context.WithCancel(context.Background())
	store := &cancelOnFinalResolveStore{address: address, cancel: cancel}
	checker := &sessionCheckerStub{results: []sessionCheckStubResult{{active: true}}}
	authorizer, _ := NewAddressAuthorizer(store, checker)
	resolved, allowed, err := authorizer.Authorize(ctx, address.InstallationID, address.BindingID, address.UserID)
	if allowed || resolved.ProviderToken != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("Authorize() = %#v, %t, %v", resolved, allowed, err)
	}
}

type cancelOnFinalResolveStore struct {
	address ResolvedAddress
	cancel  context.CancelFunc
	reads   int
}

func (s *cancelOnFinalResolveStore) ResolveActiveAddress(context.Context, string, string, uint64) (ResolvedAddress, bool, error) {
	s.reads++
	if s.reads == 2 {
		s.cancel()
	}
	return s.address, true, nil
}

func (*cancelOnFinalResolveStore) DetachIfCurrent(context.Context, Binding) (bool, error) {
	return true, nil
}

func testResolvedAddress(version, token string) ResolvedAddress {
	return ResolvedAddress{
		Binding: Binding{
			InstallationID: "install-admission-test",
			BindingID:      "d0000000-0000-4000-8000-000000000070",
			UserID:         77,
			SessionID:      testSessionID(70),
			Version:        version,
		},
		Platform: "android", Provider: ProviderFCM, ProviderProject: "project",
		Environment: "test", ProviderToken: token,
	}
}
