package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestEnvironmentBindingLoadsInfrastructureSettingsWithoutDotEnv(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Chdir(t.TempDir())

	for key, value := range map[string]string{
		"APP_ENV":                       "DEV",
		"ENABLE_STARTUP_SQL_MIGRATIONS": "true",
		"PORT":                          "8080",
		"DB_HOST":                       "postgres",
		"DB_PORT":                       "5432",
		"DB_USER":                       "friend_sheep",
		"DB_PASSWORD":                   "local-password",
		"DB_NAME":                       "friend_sheep",
		"MONGO_URI":                     "mongodb://mongo:27017/",
		"MONGO_DB":                      "friend_sheep",
		"REDIS_ADDR":                    "redis:6379",
		"REDIS_PASS":                    "redis-password",
		"REDIS_DB":                      "2",
		"SMTP_EMAIL":                    "noreply@example.test",
		"SMTP_USERNAME":                 "smtp-user",
		"SMTP_PASSWORD":                 "smtp-password",
		"SMTP_HOST":                     "smtp",
		"SMTP_PORT":                     "2525",
		"S3_ACCESS_KEY":                 "access-key",
		"S3_SECRET_KEY":                 "secret-key",
		"S3_ENDPOINT":                   "http://minio:9000",
		"S3_REGION":                     "local",
		"S3_CONTID":                     "local-container",
		"S3_BUCKET":                     "friend-sheep",
		"SECRET_KEY_JWT":                "0123456789abcdef0123456789abcdef",
		"JWT_KEY_ID":                    "local-v1",
		"NOTIFY_SERVICE_TOKEN":          "local-notify-token",
		"NOTIFY_SERVICE_BASE_URL":       "http://notify_service:8080",
	} {
		t.Setenv(key, value)
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got.AppEnv != "DEV" || !got.EnableStartupSQLMigrations || got.ServerPort != "8080" {
		t.Fatalf("общие настройки из окружения не загружены: %#v", got)
	}
	if got.Postgres != (PostgresConfig{Host: "postgres", Port: "5432", User: "friend_sheep", Password: "local-password", DbName: "friend_sheep"}) {
		t.Fatalf("Postgres из окружения = %#v", got.Postgres)
	}
	if got.Mongo != (MongoConfig{URI: "mongodb://mongo:27017/", DB: "friend_sheep"}) {
		t.Fatalf("Mongo из окружения = %#v", got.Mongo)
	}
	if got.Redis != (RedisConfig{Addr: "redis:6379", Password: "redis-password", DB: 2}) {
		t.Fatalf("Redis из окружения = %#v", got.Redis)
	}
	if got.Email != (EmailConfig{From: "noreply@example.test", Username: "smtp-user", Password: "smtp-password", SmtpHost: "smtp", SmtpPort: "2525"}) {
		t.Fatalf("SMTP из окружения = %#v", got.Email)
	}
	if got.S3 != (S3Storage{AccessKey: "access-key", SecretKey: "secret-key", Endpoint: "http://minio:9000", Region: "local", ContainerId: "local-container", Bucket: "friend-sheep"}) {
		t.Fatalf("S3 из окружения = %#v", got.S3)
	}
}
