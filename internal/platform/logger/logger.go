package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		parsedLevel = zerolog.InfoLevel
	}

	return logger.Level(parsedLevel)
}
