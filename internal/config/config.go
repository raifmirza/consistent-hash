package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig    `yaml:"server"`
	HashRing HashRingConfig  `yaml:"hashRing"`
	Health   HealthConfig    `yaml:"health"`
	Backends []BackendConfig `yaml:"backends"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type HashRingConfig struct {
	Replicas int `yaml:"replicas"`
}

type HealthConfig struct {
	Interval         time.Duration `yaml:"interval"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failureThreshold"`
	SuccessThreshold int           `yaml:"successThreshold"`
}

type BackendConfig struct {
	ID      string `yaml:"id"`
	Address string `yaml:"address"`
}

func Load(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(f, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.HashRing.Replicas <= 0 {
		return fmt.Errorf("hashRing.replicas must be greater than zero")
	}

	if c.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}

	if len(c.Backends) == 0 {
		return fmt.Errorf("at least one backend is required")
	}

	return nil
}
