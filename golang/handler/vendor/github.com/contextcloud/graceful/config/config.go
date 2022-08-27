package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type TracingConfig struct {
	Enabled bool
	Type    string
	Url     string
}

type Config struct {
	v *viper.Viper

	Environment string
	ServiceName string
	Version     string
	SrvAddr     string
	MetricsAddr string
	HealthAddr  string

	Tracing TracingConfig
}

func (c *Config) Parse(cfg interface{}) error {
	if err := c.v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

func NewConfig(ctx context.Context) (Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AddConfigPath("./config")
	v.SetConfigName("default")
	v.AutomaticEnv()

	cfg := Config{
		v:           v,
		Environment: "production",
		ServiceName: "service",
		Version:     "1.0.0",
		SrvAddr:     ":8080",
		MetricsAddr: ":8081",
		HealthAddr:  ":8082",
	}

	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal base config: %w", err)
	}

	return cfg, nil
}
