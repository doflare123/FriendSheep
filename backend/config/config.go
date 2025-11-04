package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv       string      `mapstructure:"APP_ENV"`
	ServerPort   string      `mapstructure:"PORT"`
	JWTSecretKey string      `mapstructure:"SECRET_KEY_JWT"`
	LogLevel     string      `mapstructure:"LOG_LEVEL"`
	Email        EmailConfig `mapstructure:",squash"`

	Postgres PostgresConfig `mapstructure:",squash"`
	Mongo    MongoConfig    `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
}

type EmailConfig struct {
	From     string `mapstructure:"SMTP_HOST"`
	Password string `mapstructure:"SMTP_PORT"`
	SmtpHost string `mapstructure:"SMTP_EMAIL"`
	SmtpPort string `mapstructure:"SMTP_PASSWORD"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     string `mapstructure:"POSTGRES_PORT"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	DbName   string `mapstructure:"POSTGRES_DB"`
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

func NewConfig() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠️  .env not exist")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error config: %v", err)
	}

	return &cfg
}
