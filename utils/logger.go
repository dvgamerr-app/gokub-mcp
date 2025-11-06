package utils

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logFormat := strings.ToLower(os.Getenv("LOG_FORMAT"))
	var output io.Writer = os.Stdout
	if logFormat == "text" {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level zerolog.Level
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}
