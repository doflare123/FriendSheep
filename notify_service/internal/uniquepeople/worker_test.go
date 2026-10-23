package uniquepeople

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClient struct {
	jobs          []Claim
	execute       func(context.Context, Claim) (ExecuteResult, error)
	claimLimit    int
	executeCalled atomic.Int32
}

func (c *fakeClient) Claim(_ context.Context, limit int) (ClaimResponse, error) {
	c.claimLimit = limit
	return ClaimResponse{SchemaVersion: 1, Jobs: append([]Claim(nil), c.jobs...)}, nil
}

func (c *fakeClient) Execute(ctx context.Context, claim Claim) (ExecuteResult, error) {
	c.executeCalled.Add(1)
	return c.execute(ctx, claim)
}

func TestWorkerBoundsParallelExecutionByConfiguredConcurrency(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	client := &fakeClient{jobs: []Claim{
		{UserID: 1, Generation: 1, LeaseToken: "a", LeaseUntil: now.Add(time.Minute)},
		{UserID: 2, Generation: 1, LeaseToken: "b", LeaseUntil: now.Add(time.Minute)},
		{UserID: 3, Generation: 1, LeaseToken: "c", LeaseUntil: now.Add(time.Minute)},
		{UserID: 4, Generation: 1, LeaseToken: "d", LeaseUntil: now.Add(time.Minute)},
		{UserID: 5, Generation: 1, LeaseToken: "e", LeaseUntil: now.Add(time.Minute)},
	}}
	var active atomic.Int32
	var maximum atomic.Int32
	release := make(chan struct{})
	started := make(chan struct{}, len(client.jobs))
	client.execute = func(_ context.Context, claim Claim) (ExecuteResult, error) {
		current := active.Add(1)
		for current > maximum.Load() && !maximum.CompareAndSwap(maximum.Load(), current) {
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		return ExecuteResult{SchemaVersion: 1, UserID: claim.UserID, Generation: claim.Generation, Outcome: OutcomeApplied}, nil
	}

	worker := NewWorker(slog.New(slog.NewTextHandler(io.Discard, nil)), client, fixedClock{now}, 2, time.Second)
	done := make(chan error, 1)
	go func() {
		_, err := worker.ProcessBatch(context.Background())
		done <- err
	}()
	for range 2 {
		<-started
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if client.claimLimit != 2 || maximum.Load() != 2 || client.executeCalled.Load() != int32(len(client.jobs)) {
		t.Fatalf("claim limit=%d max parallel=%d execute calls=%d", client.claimLimit, maximum.Load(), client.executeCalled.Load())
	}
}

func TestWorkerDoesNotClaimNextBatchBeforeExecutionsFinish(t *testing.T) {
	now := time.Now().UTC()
	started := make(chan struct{})
	client := &fakeClient{jobs: []Claim{{UserID: 1, Generation: 1, LeaseToken: "a", LeaseUntil: now.Add(time.Minute)}}}
	client.execute = func(ctx context.Context, claim Claim) (ExecuteResult, error) {
		close(started)
		<-ctx.Done()
		return ExecuteResult{}, ctx.Err()
	}
	worker := NewWorker(slog.New(slog.NewTextHandler(io.Discard, nil)), client, fixedClock{now}, 1, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := worker.ProcessBatch(ctx)
		done <- err
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("ProcessBatch() error = %v", err)
	}
	if client.executeCalled.Load() != 1 {
		t.Fatalf("execute calls = %d", client.executeCalled.Load())
	}
}

func TestManagerShutdownCancelsInflightExecute(t *testing.T) {
	now := time.Now().UTC()
	started := make(chan struct{})
	var once sync.Once
	client := &fakeClient{jobs: []Claim{{UserID: 1, Generation: 1, LeaseToken: "a", LeaseUntil: now.Add(time.Minute)}}}
	client.execute = func(ctx context.Context, _ Claim) (ExecuteResult, error) {
		once.Do(func() { close(started) })
		<-ctx.Done()
		return ExecuteResult{}, ctx.Err()
	}
	manager := &Manager{
		worker: NewWorker(slog.New(slog.NewTextHandler(io.Discard, nil)), client, fixedClock{now}, 1, time.Hour),
		done:   make(chan struct{}),
	}
	if err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-started
	if err := manager.Stop(time.Second); err != nil {
		t.Fatal(err)
	}
}
