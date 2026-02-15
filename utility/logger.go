package utility

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// InitLogger initializes the global logger with the specified configuration
func InitLogger(logLevel string) {
	// Parse log level
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	// Create console writer with color and pretty formatting for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Use multi-writer to log to both console and file if needed
	multi := io.MultiWriter(consoleWriter)

	Logger = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set as global logger
	log.Logger = Logger

	Logger.Info().
		Str("level", logLevel).
		Msg("Logger initialized")
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &Logger
}
