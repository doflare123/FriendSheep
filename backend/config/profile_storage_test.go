package config

import (
	"strings"
	"testing"
	"time"
)

func validProfileStorageConfig() S3Storage {
	return S3Storage{Enabled: true, AccessKey: "access", SecretKey: "secret", Region: "ru-1", Bucket: "avatars", Endpoint: "https://s3.ru-1.storage.selcloud.ru/", PublicBaseURL: "https://bucket.selstorage.ru/"}
}

func TestNormalizeProfileStorage(t *testing.T) {
	cfg := Config{}
	if err := cfg.NormalizeProfileStorage(); err != nil || cfg.S3.HTTPTimeout != 0 {
		t.Fatalf("disabled = %+v, %v", cfg.S3, err)
	}
	cfg.S3 = validProfileStorageConfig()
	if err := cfg.NormalizeProfileStorage(); err != nil {
		t.Fatal(err)
	}
	if cfg.S3.HTTPTimeout != 15*time.Second || cfg.S3.Endpoint != "https://s3.ru-1.storage.selcloud.ru" || cfg.S3.PublicBaseURL != "https://bucket.selstorage.ru" {
		t.Fatalf("normalized = %+v", cfg.S3)
	}
}

func TestNormalizeProfileStorageRejectsInvalid(t *testing.T) {
	for name, mutate := range map[string]func(*S3Storage){
		"access":          func(c *S3Storage) { c.AccessKey = "" },
		"secret":          func(c *S3Storage) { c.SecretKey = "" },
		"region":          func(c *S3Storage) { c.Region = "" },
		"bucket":          func(c *S3Storage) { c.Bucket = "../other" },
		"bucket_upper":    func(c *S3Storage) { c.Bucket = "Avatars" },
		"bucket_symbol":   func(c *S3Storage) { c.Bucket = "bad_bucket" },
		"bucket_unicode":  func(c *S3Storage) { c.Bucket = "аватары" },
		"bucket_colon":    func(c *S3Storage) { c.Bucket = "bucket:other" },
		"bucket_length":   func(c *S3Storage) { c.Bucket = strings.Repeat("a", 129) },
		"bucket_short":    func(c *S3Storage) { c.Bucket = "a" },
		"bucket_ip":       func(c *S3Storage) { c.Bucket = "127.0.0.1" },
		"endpoint":        func(c *S3Storage) { c.Endpoint = "http://s3.example.com" },
		"endpoint_user":   func(c *S3Storage) { c.Endpoint = "https://user:pass@s3.example.com" },
		"endpoint_path":   func(c *S3Storage) { c.Endpoint += "bucket" },
		"endpoint_query":  func(c *S3Storage) { c.Endpoint += "?x=1" },
		"public_empty":    func(c *S3Storage) { c.PublicBaseURL = "" },
		"public_http":     func(c *S3Storage) { c.PublicBaseURL = "http://localhost:9000" },
		"public_user":     func(c *S3Storage) { c.PublicBaseURL = "https://user@bucket.selstorage.ru" },
		"public_fragment": func(c *S3Storage) { c.PublicBaseURL += "#x" },
		"timeout":         func(c *S3Storage) { c.HTTPTimeout = -time.Second },
		"timeout_large":   func(c *S3Storage) { c.HTTPTimeout = 2 * time.Minute },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validProfileStorageConfig()
			mutate(&cfg)
			if err := cfg.NormalizeProfileStorage(); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
	for _, endpoint := range []string{"http://localhost:9000", "http://127.0.0.1:9000", "http://[::1]:9000"} {
		cfg := validProfileStorageConfig()
		cfg.Endpoint = endpoint
		if err := cfg.NormalizeProfileStorage(); err != nil {
			t.Errorf("loopback %s: %v", endpoint, err)
		}
	}
}
