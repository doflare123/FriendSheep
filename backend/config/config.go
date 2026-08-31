package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv                     string          `mapstructure:"APP_ENV"`
	EnableStartupSQLMigrations bool            `mapstructure:"ENABLE_STARTUP_SQL_MIGRATIONS"`
	ServerPort                 string          `mapstructure:"PORT"`
	JWTSecretKey               string          `mapstructure:"SECRET_KEY_JWT"`
	JWTKeyID                   string          `mapstructure:"JWT_KEY_ID"`
	JWTPreviousSecretKey       string          `mapstructure:"JWT_PREVIOUS_SECRET_KEY"`
	JWTPreviousKeyID           string          `mapstructure:"JWT_PREVIOUS_KEY_ID"`
	NotifyServiceToken         string          `mapstructure:"NOTIFY_SERVICE_TOKEN"`
	NotifyServiceBaseURL       string          `mapstructure:"NOTIFY_SERVICE_BASE_URL"`
	NotifyServiceHTTPTimeout   time.Duration   `mapstructure:"NOTIFY_SERVICE_HTTP_TIMEOUT"`
	Auth                       AuthConfig      `mapstructure:",squash"`
	LogLevel                   string          `mapstructure:"LOG_LEVEL"`
	HTTP                       HTTPConfig      `mapstructure:",squash"`
	RateLimit                  RateLimitConfig `mapstructure:",squash"`
	Email                      EmailConfig     `mapstructure:",squash"`

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
	AccessKey   string `mapstructure:"S3_ACCESS_KEY"`
	SecretKey   string `mapstructure:"S3_SECRET_KEY"`
	Endpoint    string `mapstructure:"S3_ENDPOINT"`
	Region      string `mapstructure:"S3_REGION"`
	ContainerId string `mapstructure:"S3_CONTID"`
	Bucket      string `mapstructure:"S3_BUCKET"`
}

type Upload struct {
	MaxImageSize int `mapstructure:"MAX_IMAGE_SIZE"`
}

func NewConfig() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	applyConfigDefaults()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠️  .env not exist")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error config: %v", err)
	}
	if err := cfg.NormalizeAndValidate(); err != nil {
		log.Fatalf("Error config: %v", err)
	}

	return &cfg
}
