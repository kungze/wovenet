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
	var config Config
	err := viper.UnmarshalKey("logger", &config)
	if err != nil {
		return err
	}

	err = CheckAndSetDefaultConfig(&config)
	if err != nil {
		return err
	}

	var output io.Writer
	if config.File != "" {
		output, err = os.OpenFile(config.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
	} else {
		output = os.Stdout
	}

	var programLevel = new(slog.LevelVar)
	var logger *slog.Logger
	switch strings.ToLower(config.Format) {
	case JSON:
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: programLevel}))
	case TEXT:
		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: programLevel}))
	default:
		return fmt.Errorf("unsupported log format")
	}
	slog.SetDefault(logger)
	switch strings.ToUpper(config.Level) {
	case INFO:
		programLevel.Set(slog.LevelInfo)
	case WARN:
		programLevel.Set(slog.LevelWarn)
	case ERROR:
		programLevel.Set(slog.LevelError)
	case DEBUG:
		programLevel.Set(slog.LevelDebug)
	default:
		return fmt.Errorf("unsupported log level: %s", config.Level)
	}
	return nil
}

func GetDefault() *slog.Logger {
	return slog.Default()
}
