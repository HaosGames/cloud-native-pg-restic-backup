package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger to provide structured logging
type Logger struct {
	zerolog.Logger
}

// Config holds logging configuration
type Config struct {
	Level      string `json:"level"`
	JSONOutput bool   `json:"jsonOutput"`
}

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg Config) *Logger {
	// Set default level if not specified
	level := zerolog.InfoLevel
	if cfg.Level != "" {
		if l, err := zerolog.ParseLevel(cfg.Level); err == nil {
			level = l
		}
	}

	// Configure output format
	var output zerolog.ConsoleWriter
	if cfg.JSONOutput {
		return &Logger{
			Logger: zerolog.New(os.Stdout).
				Level(level).
				With().
				Timestamp().
				Logger(),
		}
	}

	output = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	return &Logger{
		Logger: zerolog.New(output).
			Level(level).
			With().
			Timestamp().
			Logger(),
	}
}

// Component adds a component field to the logger
func (l *Logger) Component(component string) *Logger {
	return &Logger{
		Logger: l.With().Str("component", component).Logger(),
	}
}

// Operation adds an operation field to the logger
func (l *Logger) Operation(operation string) *Logger {
	return &Logger{
		Logger: l.With().Str("operation", operation).Logger(),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		Logger: ctx.Logger(),
	}
}

// Printf implements the plugin.Logger interface
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info().Msgf(format, v...)
}
