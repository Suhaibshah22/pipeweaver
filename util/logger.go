package util

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Logger is the singleton instance of slog.Logger
var (
	instance *slog.Logger
	once     sync.Once
)

// InitLogger initializes the slog logger with JSON formatting and sets the log level.
// It should be called once, typically in the dependencies initialization.
func InitLogger(logLevel string, environment string) {
	once.Do(func() {
		level, err := parseLogLevel(logLevel)
		if err != nil {
			level = slog.LevelDebug // Default level
		}

		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})

		instance = slog.New(handler)

		// Add global context fields
		instance = instance.With(
			"sdk", "slog",
			"app", "pipeweaver",
			"environment", environment,
		)

		instance.Debug("Logger initialized", "level", level.String())
	})
}

// GetLogger returns the singleton logger instance.
// It must be initialized using InitLogger before calling.
func GetLogger() *slog.Logger {
	if instance == nil {
		InitLogger("debug", "production")
	}
	return instance
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
