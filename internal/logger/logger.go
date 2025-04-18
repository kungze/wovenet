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
	cfg := &Config{}
	err := viper.UnmarshalKey("logger", cfg)
	if err != nil {
		return err
	}

	cfg, err = CheckAndSetDefaultConfig(*cfg)
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
	case JSON:
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: programLevel}))
	case TEXT:
		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: programLevel}))
	default:
		return fmt.Errorf("unsupported log format")
	}
	slog.SetDefault(logger)
	switch strings.ToUpper(cfg.Level) {
	case INFO:
		programLevel.Set(slog.LevelInfo)
	case WARN:
		programLevel.Set(slog.LevelWarn)
	case ERROR:
		programLevel.Set(slog.LevelError)
	case DEBUG:
		programLevel.Set(slog.LevelDebug)
	default:
		return fmt.Errorf("unsupported log level: %s", cfg.Level)
	}
	return nil
}

func GetDefault() *slog.Logger {
	return slog.Default()
}
