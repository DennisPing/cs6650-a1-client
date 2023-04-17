package log

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

var Logger zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // Set default log level to INFO

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err == nil {
			zerolog.SetGlobalLevel(level)
		}
	}

	// fmt.Printf("Current log level: %s\n", zerolog.GlobalLevel().String())
	wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages", missed)
	})

	logger := zerolog.New(wr).With().Timestamp().Logger()

	// Include code location if log level is DEBUG
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		logger = logger.With().Caller().Logger()
	} else {
		logger = logger.With().Logger()
	}

	// Set the logger with the proper context to the global Logger
	Logger = logger
}
