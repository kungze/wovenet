package logger

import (
	"fmt"
	"slices"
	"strings"
)

type Config struct {
	Level  string `mapstructure:"level"`
	File   string `mapstructure:"file"`
	Format string `mapstructure:"format"`
}

func CheckAndSetDefaultConfig(config Config) (*Config, error) {
	if config.Level == "" {
		config.Level = INFO
	}
	if config.Format == "" {
		config.Format = TEXT
	}
	if !slices.Contains([]string{INFO, WARN, ERROR, DEBUG}, strings.ToUpper(config.Level)) {
		return nil, fmt.Errorf("unsupported log level: %s", config.Level)
	}
	if !slices.Contains([]string{JSON, TEXT}, strings.ToLower(config.Format)) {
		return nil, fmt.Errorf("unsupported log format: %s", config.Format)
	}
	return &config, nil
}
