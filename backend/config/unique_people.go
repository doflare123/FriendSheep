package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// UniquePeopleConfig ограничивает фоновые запросы ко всему deployment монолита.
type UniquePeopleConfig struct {
	ClaimBatch    int           `mapstructure:"STATISTICS_UNIQUE_CLAIM_BATCH"`
	Concurrency   int           `mapstructure:"STATISTICS_UNIQUE_CONCURRENCY"`
	QueryTimeout  time.Duration `mapstructure:"STATISTICS_UNIQUE_QUERY_TIMEOUT"`
	LeaseDuration time.Duration `mapstructure:"STATISTICS_UNIQUE_LEASE_DURATION"`
	RetryBase     time.Duration `mapstructure:"STATISTICS_UNIQUE_RETRY_BASE"`
	RetryMax      time.Duration `mapstructure:"STATISTICS_UNIQUE_RETRY_MAX"`
}

func DefaultUniquePeopleConfig() UniquePeopleConfig {
	return UniquePeopleConfig{ClaimBatch: 1, Concurrency: 1, QueryTimeout: 20 * time.Second, LeaseDuration: 2 * time.Minute, RetryBase: time.Minute, RetryMax: time.Hour}
}

func (c UniquePeopleConfig) Validate() error {
	var errs []error
	if c.ClaimBatch < 1 || c.ClaimBatch > 8 {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_CLAIM_BATCH: ожидается 1..8"))
	}
	if c.Concurrency < 1 || c.Concurrency > 8 {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_CONCURRENCY: ожидается 1..8"))
	}
	if c.QueryTimeout < time.Millisecond || c.QueryTimeout > 30*time.Second {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_QUERY_TIMEOUT: ожидается 1ms..30s"))
	}
	if c.LeaseDuration <= c.QueryTimeout+5*time.Second || c.LeaseDuration > 10*time.Minute {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_LEASE_DURATION: требуется запас 5s сверх query timeout, максимум 10m"))
	}
	if c.RetryBase < time.Second || c.RetryBase > time.Hour {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_RETRY_BASE: ожидается 1s..1h"))
	}
	if c.RetryMax < c.RetryBase || c.RetryMax > 6*time.Hour {
		errs = append(errs, fmt.Errorf("STATISTICS_UNIQUE_RETRY_MAX: требуется base..6h"))
	}
	return errors.Join(errs...)
}

func applyUniquePeopleDefaults() {
	c := DefaultUniquePeopleConfig()
	viper.SetDefault("STATISTICS_UNIQUE_CLAIM_BATCH", c.ClaimBatch)
	viper.SetDefault("STATISTICS_UNIQUE_CONCURRENCY", c.Concurrency)
	viper.SetDefault("STATISTICS_UNIQUE_QUERY_TIMEOUT", c.QueryTimeout.String())
	viper.SetDefault("STATISTICS_UNIQUE_LEASE_DURATION", c.LeaseDuration.String())
	viper.SetDefault("STATISTICS_UNIQUE_RETRY_BASE", c.RetryBase.String())
	viper.SetDefault("STATISTICS_UNIQUE_RETRY_MAX", c.RetryMax.String())
}
