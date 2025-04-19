package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strconv"
	"time"
)

type LevelWriter struct {
	io.Writer
	Levels []zerolog.Level
}

func (w LevelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	for _, l := range w.Levels {
		if l == level {
			return w.Write(p)
		}
	}
	return len(p), nil
}

func newLevelWriter() zerolog.LevelWriter {
	writer := zerolog.MultiLevelWriter(
		LevelWriter{
			Writer: zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
			Levels: []zerolog.Level{
				zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
			},
		},
		LevelWriter{
			Writer: zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
			Levels: []zerolog.Level{
				zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
			},
		},
	)
	return writer
}

func Init() {
	logger := zerolog.New(newLevelWriter()).With().Timestamp().Logger()
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

	//nolint:reassign //not passed around
	log.Logger = logger
}
