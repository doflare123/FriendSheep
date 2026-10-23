package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestUniquePeopleConfigBounds(t *testing.T) {
	if err := DefaultUniquePeopleConfig().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*UniquePeopleConfig){
		func(c *UniquePeopleConfig) { c.ClaimBatch = 0 },
		func(c *UniquePeopleConfig) { c.ClaimBatch = 9 },
		func(c *UniquePeopleConfig) { c.Concurrency = 0 },
		func(c *UniquePeopleConfig) { c.Concurrency = 9 },
		func(c *UniquePeopleConfig) { c.QueryTimeout = 0 },
		func(c *UniquePeopleConfig) { c.QueryTimeout = 31 * time.Second },
		func(c *UniquePeopleConfig) { c.LeaseDuration = c.QueryTimeout },
		func(c *UniquePeopleConfig) { c.RetryBase = 0 },
		func(c *UniquePeopleConfig) { c.RetryMax = c.RetryBase - time.Second },
	} {
		c := DefaultUniquePeopleConfig()
		change(&c)
		if c.Validate() == nil {
			t.Errorf("допущена некорректная конфигурация: %+v", c)
		}
	}
}

func TestUniquePeopleEnvironmentDefaultsAndOverride(t *testing.T) {
	resetViperForConfigTest(t)
	applyConfigDefaults()
	viper.AutomaticEnv()
	t.Setenv("STATISTICS_UNIQUE_CONCURRENCY", "2")
	t.Setenv("STATISTICS_UNIQUE_QUERY_TIMEOUT", "15s")
	// Проверяем чтение через тот же Viper, который использует startup.
	c := Config{}
	if err := viper.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}
	if c.UniquePeople.Concurrency != 2 || c.UniquePeople.QueryTimeout != 15*time.Second || c.UniquePeople.ClaimBatch != 1 {
		t.Fatalf("config=%+v", c.UniquePeople)
	}
}
