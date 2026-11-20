package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"friendship/config"
)

func profileStorageConfig(endpoint string) config.S3Storage {
	return config.S3Storage{Enabled: true, AccessKey: "test-access", SecretKey: "test-secret", Region: "ru-1", Bucket: "avatars", Endpoint: endpoint, PublicBaseURL: "https://bucket.selstorage.ru/base", HTTPTimeout: time.Second}
}

func TestSelectelProfileStoragePutDeleteContract(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/avatars/users/7/avatar.png" {
			t.Errorf("path = %q", r.URL.Path)
		}
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") || !strings.Contains(authorization, "Credential=test-access/") || !strings.Contains(authorization, "/ru-1/s3/aws4_request") {
			t.Errorf("signature = %q", authorization)
		}
		if r.Header.Get("X-Amz-Acl") != "" || r.Header.Get("X-Amz-Sdk-Checksum-Algorithm") != "" {
			t.Errorf("unexpected ACL/checksum headers: %v", r.Header)
		}
		switch r.Method {
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil || string(data) != "image-content" || r.Header.Get("Content-Type") != "image/png" {
				t.Errorf("body/type = %q/%q, err=%v", data, r.Header.Get("Content-Type"), err)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("method = %s", r.Method)
		}
	}))
	defer server.Close()
	storage, err := NewSelectelProfileStorage(profileStorageConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	got, err := storage.Put(context.Background(), "users/7/avatar.png", []byte("image-content"), "image/png")
	if err != nil || got != "https://bucket.selstorage.ru/base/users/7/avatar.png" {
		t.Fatalf("Put = %q, %v", got, err)
	}
	if err := storage.Delete(context.Background(), "users/7/avatar.png"); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestSelectelProfileStorageRejectsUnsafeKeysBeforeHTTP(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	storage, err := NewSelectelProfileStorage(profileStorageConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "../other", "/users/1/x", "users/../x", "users//x", "users\\x", "https://example.com/x", "users/%2e%2e/x", "users/x?other", "users/x\n"} {
		if _, err := storage.Put(context.Background(), key, nil, "image/png"); err == nil {
			t.Errorf("Put accepted %q", key)
		}
		if err := storage.Delete(context.Background(), key); err == nil {
			t.Errorf("Delete accepted %q", key)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("unsafe requests = %d", calls.Load())
	}
}

func TestSelectelProfileStorageCancellationAndFailureBound(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "<Error><Code>ServiceUnavailable</Code></Error>")
	}))
	defer server.Close()
	cfg := profileStorageConfig(server.URL)
	cfg.HTTPTimeout = 150 * time.Millisecond
	storage, err := NewSelectelProfileStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := storage.Put(ctx, "users/1/x.png", nil, "image/png"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("canceled request reached server")
	}
	started := time.Now()
	if _, err := storage.Put(context.Background(), "users/1/x.png", nil, "image/png"); err == nil {
		t.Fatal("server failure succeeded")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("request exceeded deadline: %v", elapsed)
	}
	if calls.Load() > 2 {
		t.Fatalf("attempt count = %d", calls.Load())
	}
}

func TestSelectelProfileStorageRejectsRedirect(t *testing.T) {
	var leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { leaked.Add(1) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	storage, err := NewSelectelProfileStorage(profileStorageConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Put(context.Background(), "users/1/x.png", nil, "image/png"); err == nil {
		t.Fatal("redirect succeeded")
	}
	if leaked.Load() != 0 {
		t.Fatal("redirect forwarded signed request")
	}
}
