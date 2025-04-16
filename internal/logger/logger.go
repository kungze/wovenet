package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitLogging() error {
	var cfg Config
	err := viper.UnmarshalKey("logger", &cfg)
	if err != nil {
		return err
	}

	var output io.Writer
	if cfg.File != "" {
		output, err = os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
	} else {
		output = os.Stdout
	}

	var programLevel = new(slog.LevelVar)
	var logger *slog.Logger
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: programLevel}))
	default:
		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: programLevel}))
	}
	slog.SetDefault(logger)
	switch strings.ToLower(cfg.Level) {
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	case "debug":
		programLevel.Set(slog.LevelDebug)
	default:
		return fmt.Errorf("unsupported log level: %s", cfg.Level)
	}
	return nil
}

func GetDefault() *slog.Logger {
	return slog.Default()
}
