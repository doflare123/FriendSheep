package group

import (
	"context"
	"errors"
	"testing"

	"friendship/repository"
)

func TestGORMGroupRepositoryWithinTransactionForwardsContext(t *testing.T) {
	store := &contextTransactionRepositoryStub{}
	uow := NewGORMGroupRepository(store)

	type contextKey string
	const key contextKey = "group-uow-context"
	ctx := context.WithValue(context.Background(), key, "request-value")
	writeCalled := false

	err := uow.WithinTransaction(ctx, func(groupTx) error {
		writeCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTransaction returned error: %v", err)
	}
	if store.seenCtx != ctx {
		t.Fatal("transaction runner did not receive the exact caller context")
	}
	if got := store.seenCtx.Value(key); got != "request-value" {
		t.Fatalf("transaction context value = %v, want request-value", got)
	}
	if !store.transactionCallbackCalled {
		t.Fatal("repository transaction callback was not called")
	}
	if !writeCalled {
		t.Fatal("group write callback was not called")
	}
	if !store.committed {
		t.Fatal("transaction was not committed")
	}
}

func TestGORMGroupRepositoryWithinTransactionCanceledContextSkipsWriteAndCommit(t *testing.T) {
	store := &contextTransactionRepositoryStub{}
	uow := NewGORMGroupRepository(store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	writeCalled := false

	err := uow.WithinTransaction(ctx, func(groupTx) error {
		writeCalled = true
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithinTransaction error = %v, want context.Canceled", err)
	}
	if store.seenCtx != ctx {
		t.Fatal("transaction runner did not receive the exact canceled context")
	}
	if store.transactionCallbackCalled {
		t.Fatal("repository transaction callback ran for an already-canceled context")
	}
	if writeCalled {
		t.Fatal("group write callback ran for an already-canceled context")
	}
	if store.committed {
		t.Fatal("transaction committed for an already-canceled context")
	}
}

type contextTransactionRepositoryStub struct {
	repository.PostgresRepository
	seenCtx                   context.Context
	transactionCallbackCalled bool
	committed                 bool
}

func (s *contextTransactionRepositoryStub) TransactionWithContext(
	ctx context.Context,
	fn func(repository.PostgresRepository) error,
) error {
	s.seenCtx = ctx
	if err := ctx.Err(); err != nil {
		return err
	}

	s.transactionCallbackCalled = true
	if err := fn(s); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.committed = true
	return nil
}
