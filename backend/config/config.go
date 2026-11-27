package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv                     string             `mapstructure:"APP_ENV"`
	EnableStartupSQLMigrations bool               `mapstructure:"ENABLE_STARTUP_SQL_MIGRATIONS"`
	ServerPort                 string             `mapstructure:"PORT"`
	JWTSecretKey               string             `mapstructure:"SECRET_KEY_JWT"`
	JWTKeyID                   string             `mapstructure:"JWT_KEY_ID"`
	JWTPreviousSecretKey       string             `mapstructure:"JWT_PREVIOUS_SECRET_KEY"`
	JWTPreviousKeyID           string             `mapstructure:"JWT_PREVIOUS_KEY_ID"`
	NotifyServiceToken         string             `mapstructure:"NOTIFY_SERVICE_TOKEN"`
	NotifyServiceBaseURL       string             `mapstructure:"NOTIFY_SERVICE_BASE_URL"`
	NotifyServiceHTTPTimeout   time.Duration      `mapstructure:"NOTIFY_SERVICE_HTTP_TIMEOUT"`
	NotifyPushProviderProject  string             `mapstructure:"NOTIFY_PUSH_PROVIDER_PROJECT"`
	NotifyPushEnvironment      string             `mapstructure:"NOTIFY_PUSH_ENVIRONMENT"`
	UniquePeople               UniquePeopleConfig `mapstructure:",squash"`
	Auth                       AuthConfig         `mapstructure:",squash"`
	LogLevel                   string             `mapstructure:"LOG_LEVEL"`
	HTTP                       HTTPConfig         `mapstructure:",squash"`
	RateLimit                  RateLimitConfig    `mapstructure:",squash"`
	Email                      EmailConfig        `mapstructure:",squash"`

	Postgres PostgresConfig `mapstructure:",squash"`
	Mongo    MongoConfig    `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`

	S3 S3Storage `mapstructure:",squash"`

	Upload Upload `mapstructure:",squash"`
}

type AuthConfig struct {
	Issuer          string        `mapstructure:"JWT_ISSUER"`
	Audience        string        `mapstructure:"JWT_AUDIENCE"`
	AccessTokenTTL  time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTokenTTL time.Duration `mapstructure:"AUTH_REFRESH_TTL"`
	ClockSkew       time.Duration `mapstructure:"JWT_CLOCK_SKEW"`
}

func (c Config) DerivedRateLimitHashSecret() string {
	mac := hmac.New(sha256.New, []byte(c.JWTSecretKey))
	_, _ = mac.Write([]byte("friendSheep/rate-limit/v1"))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

type EmailConfig struct {
	From     string `mapstructure:"SMTP_EMAIL"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	SmtpHost string `mapstructure:"SMTP_HOST"`
	SmtpPort string `mapstructure:"SMTP_PORT"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DbName   string `mapstructure:"DB_NAME"`
}

type MongoConfig struct {
	URI string `mapstructure:"MONGO_URI"`
	DB  string `mapstructure:"MONGO_DB"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASS"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type S3Storage struct {
	Enabled       bool          `mapstructure:"S3_ENABLED"`
	PublicBaseURL string        `mapstructure:"S3_PUBLIC_BASE_URL"`
	HTTPTimeout   time.Duration `mapstructure:"S3_HTTP_TIMEOUT"`
	AccessKey     string        `mapstructure:"S3_ACCESS_KEY"`
	SecretKey     string        `mapstructure:"S3_SECRET_KEY"`
	Endpoint      string        `mapstructure:"S3_ENDPOINT"`
	Region        string        `mapstructure:"S3_REGION"`
	ContainerId   string        `mapstructure:"S3_CONTID"`
	Bucket        string        `mapstructure:"S3_BUCKET"`
}

type Upload struct {
	MaxImageSize int `mapstructure:"MAX_IMAGE_SIZE"`
}

func NewConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error config: %v", err)
	}
	return cfg
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	applyConfigDefaults()
	if err := bindConfigEnvironment(reflect.TypeOf(Config{})); err != nil {
		return nil, err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠️  .env not exist")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("разбор конфигурации: %w", err)
	}
	if err := cfg.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func bindConfigEnvironment(configType reflect.Type) error {
	if configType.Kind() == reflect.Pointer {
		configType = configType.Elem()
	}
	if configType.Kind() != reflect.Struct {
		return fmt.Errorf("тип конфигурации должен быть структурой, получено %s", configType.Kind())
	}

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		parts := strings.Split(field.Tag.Get("mapstructure"), ",")
		name := parts[0]
		if name == "-" {
			continue
		}

		squashed := false
		for _, option := range parts[1:] {
			if option == "squash" {
				squashed = true
				break
			}
		}
		if squashed {
			if err := bindConfigEnvironment(field.Type); err != nil {
				return fmt.Errorf("привязка окружения для %s: %w", field.Name, err)
			}
			continue
		}
		if name == "" {
			continue
		}
		if err := viper.BindEnv(name); err != nil {
			return fmt.Errorf("привязка переменной окружения %s: %w", name, err)
		}
	}
	return nil
}
