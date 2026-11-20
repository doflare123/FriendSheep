package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode"

	"friendship/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SelectelProfileStorage хранит объекты по серверным ключам, не разбирая URL клиента.
type SelectelProfileStorage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
	timeout       time.Duration
}

func NewSelectelProfileStorage(cfg config.S3Storage) (*SelectelProfileStorage, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("хранилище аватаров не включено")
	}
	if err := cfg.NormalizeProfileStorage(); err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		HTTPClient: &http.Client{Timeout: cfg.HTTPTimeout, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
		RetryMaxAttempts:           2,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = true
	})
	return &SelectelProfileStorage{client: client, bucket: cfg.Bucket, publicBaseURL: cfg.PublicBaseURL, timeout: cfg.HTTPTimeout}, nil
}

func (s *SelectelProfileStorage) Put(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	if err := validProfileObjectKey(key); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), Body: bytes.NewReader(body),
		ContentType: aws.String(contentType), CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", fmt.Errorf("загрузка аватара в S3: %w", err)
	}
	segments := strings.Split(key, "/")
	for i := range segments {
		segments[i] = url.PathEscape(segments[i])
	}
	return s.publicBaseURL + "/" + strings.Join(segments, "/"), nil
}

func (s *SelectelProfileStorage) Delete(ctx context.Context, key string) error {
	if err := validProfileObjectKey(key); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("удаление аватара из S3: %w", err)
	}
	return nil
}

func validProfileObjectKey(key string) error {
	if key == "" || len(key) > 1024 || path.Clean(key) != key || strings.HasPrefix(key, "/") || strings.ContainsAny(key, "\\%?#") || strings.Contains(key, ":") {
		return fmt.Errorf("некорректный ключ объекта аватара")
	}
	for _, part := range strings.Split(key, "/") {
		if part == "." || part == ".." || part == "" {
			return fmt.Errorf("некорректный ключ объекта аватара")
		}
	}
	if strings.IndexFunc(key, unicode.IsControl) >= 0 {
		return fmt.Errorf("некорректный ключ объекта аватара")
	}
	return nil
}
