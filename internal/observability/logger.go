package observability

import (
	"io"
	"log/slog"
	"os"
)

type OutputFormat string

const (
	FormatConsole OutputFormat = "console"
	FormatJSON    OutputFormat = "json"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds the logger configuration
type Config struct {
	Format OutputFormat
	Level  LogLevel
	Writer io.Writer // Optional: defaults to os.Stdout
}

// InitLogger initializes and sets the global slog logger with the specified configuration
func InitLogger(cfg Config) *slog.Logger {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	var level slog.Level
	switch cfg.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	// Create appropriate handler based on format
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.Writer, opts)
	case FormatConsole:
		handler = slog.NewTextHandler(cfg.Writer, opts)
	default:
		// Default to console format
		handler = slog.NewTextHandler(cfg.Writer, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
