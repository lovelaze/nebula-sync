package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func Init() {
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Logger = logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if debugEnv := os.Getenv("NS_DEBUG"); debugEnv != "" {
		debug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to parse boolean env NS_DEBUG")
		}

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			logger = logger.With().Caller().Logger()
		}
	}

	log.Logger = logger
}
