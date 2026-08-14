package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

const lifecycleTestTimeout = 2 * time.Second

func TestServeHTTPGracefulShutdownStopsAdmissionAndDrainsActiveRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	requestCanceled := make(chan struct{}, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		select {
		case <-releaseRequest:
			w.WriteHeader(http.StatusNoContent)
		case <-r.Context().Done():
			requestCanceled <- struct{}{}
		}
	})

	listener := newLifecycleTestListener(t)
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- serveHTTP(ctx, &http.Server{Handler: handler}, listener, time.Second)
	}()

	responseDone := make(chan error, 1)
	go func() {
		resp, err := lifecycleTestClient().Get("http://" + listener.Addr().String())
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				err = errors.New("active request returned an unexpected status")
			}
		}
		responseDone <- err
	}()

	waitLifecycleSignal(t, requestStarted, "active request did not reach the handler")
	cancel()
	waitUntilListenerRejectsConnections(t, listener.Addr().String())

	select {
	case <-requestCanceled:
		t.Fatal("active request context was canceled during the graceful drain window")
	case err := <-responseDone:
		t.Fatalf("active request ended before it was released: %v", err)
	case <-time.After(75 * time.Millisecond):
	}

	close(releaseRequest)
	if err := waitLifecycleResult(t, responseDone, "active request did not finish"); err != nil {
		t.Fatalf("active request failed while draining: %v", err)
	}
	if err := waitLifecycleResult(t, serverDone, "graceful shutdown did not return"); err != nil {
		t.Fatalf("serveHTTP returned an error after graceful shutdown: %v", err)
	}
}

func TestServeHTTPForcedShutdownCancelsRequestContextAndDoesNotHang(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
		close(requestCanceled)
	})

	listener := newLifecycleTestListener(t)
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- serveHTTP(ctx, &http.Server{Handler: handler}, listener, 50*time.Millisecond)
	}()

	requestDone := make(chan error, 1)
	go func() {
		resp, err := lifecycleTestClient().Get("http://" + listener.Addr().String())
		if resp != nil {
			resp.Body.Close()
		}
		requestDone <- err
	}()

	waitLifecycleSignal(t, requestStarted, "blocking request did not reach the handler")
	startedShutdown := time.Now()
	cancel()
	waitLifecycleSignal(t, requestCanceled, "forced shutdown did not cancel the request context")
	if err := waitLifecycleResult(t, serverDone, "forced shutdown hung"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("serveHTTP error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(startedShutdown); elapsed > time.Second {
		t.Fatalf("forced shutdown took %v, want no more than 1s", elapsed)
	}
	_ = waitLifecycleResult(t, requestDone, "client request did not unblock after forced shutdown")
}

func newLifecycleTestListener(t *testing.T) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	return listener
}

func lifecycleTestClient() *http.Client {
	return &http.Client{
		Timeout: lifecycleTestTimeout,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}

func waitUntilListenerRejectsConnections(t *testing.T, address string) {
	t.Helper()
	deadline := time.Now().Add(lifecycleTestTimeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 20*time.Millisecond)
		if err != nil {
			return
		}
		_ = conn.Close()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("listener continued accepting new connections after shutdown started")
}

func waitLifecycleSignal(t *testing.T, signal <-chan struct{}, failure string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(lifecycleTestTimeout):
		t.Fatal(failure)
	}
}

func waitLifecycleResult(t *testing.T, result <-chan error, failure string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(lifecycleTestTimeout):
		t.Fatal(failure)
		return nil
	}
}
