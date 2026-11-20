package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var profileBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,126}[a-z0-9]$`)

// NormalizeProfileStorage проверяет отдельное подключение хранилища аватаров.
func (c *Config) NormalizeProfileStorage() error {
	return c.S3.NormalizeProfileStorage()
}

func (s *S3Storage) NormalizeProfileStorage() error {
	if !s.Enabled {
		return nil
	}
	for _, field := range []struct{ name, value string }{
		{"S3_ACCESS_KEY", s.AccessKey}, {"S3_SECRET_KEY", s.SecretKey},
		{"S3_REGION", s.Region}, {"S3_BUCKET", s.Bucket},
	} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s обязателен при S3_ENABLED=true", field.name)
		}
	}
	if !profileBucketPattern.MatchString(s.Bucket) || strings.Contains(s.Bucket, "..") || net.ParseIP(s.Bucket) != nil {
		return fmt.Errorf("S3_BUCKET содержит недопустимые символы")
	}
	if err := validateStorageURL(s.Endpoint, true); err != nil {
		return fmt.Errorf("S3_ENDPOINT: %w", err)
	}
	if err := validateStorageURL(s.PublicBaseURL, false); err != nil {
		return fmt.Errorf("S3_PUBLIC_BASE_URL: %w", err)
	}
	if s.HTTPTimeout == 0 {
		s.HTTPTimeout = 15 * time.Second
	}
	if s.HTTPTimeout < 0 || s.HTTPTimeout > time.Minute {
		return fmt.Errorf("S3_HTTP_TIMEOUT должен быть больше нуля и не больше минуты")
	}
	s.Endpoint = strings.TrimRight(s.Endpoint, "/")
	s.PublicBaseURL = strings.TrimRight(s.PublicBaseURL, "/")
	return nil
}

func validateStorageURL(raw string, endpoint bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return fmt.Errorf("нужен абсолютный HTTPS URL без пароля, query и fragment")
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if !endpoint || u.Scheme != "http" || (u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return fmt.Errorf("нужен HTTPS URL; HTTP endpoint разрешён только для loopback")
		}
	}
	if endpoint && u.Path != "" && u.Path != "/" {
		return fmt.Errorf("endpoint не должен содержать путь бакета")
	}
	return nil
}
