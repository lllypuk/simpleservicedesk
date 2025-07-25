package logger

import (
	"log/slog"
	"os"
	"strings"
)

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Setup настраивает глобальный логгер
func Setup() {
	level := getLogLevel()
	format := getLogFormat()

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch format {
	case LogFormatText:
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// getLogLevel определяет уровень логирования из переменной окружения
func getLogLevel() slog.Level {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getLogFormat определяет формат логирования из переменной окружения
func getLogFormat() string {
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if format == LogFormatText {
		return LogFormatText
	}
	return LogFormatJSON
}
