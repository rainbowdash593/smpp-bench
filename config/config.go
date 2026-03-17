package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -..
	Config struct {
		Connection `yaml:"connection"`
		Message    `yaml:"message"`
		Log        `yaml:"logger"`
	}

	// Connection -..
	Connection struct {
		Host          string `yaml:"host"`
		Login         string `yaml:"login"`
		Password      string `yaml:"password"`
		WindowSize    uint   `yaml:"window_size"`
		BindTimeoutMs int    `yaml:"bind_timout_ms"`
	}

	// Message -..
	Message struct {
		PhoneCountry string `yaml:"phone_country"`
		Source       string `yaml:"source"`
		Random       bool   `yaml:"random"`
		Text         string `yaml:"text"`
	}

	// Log -..
	Log struct {
		Level string `yaml:"level"`
	}
)

// NewConfig returns app config.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
