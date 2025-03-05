package sync

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	delay                  = 1 * time.Second
	AttemptsPostTeleporter = 5
	AttemptsPatchConfig    = 5
	AttemptsPostRunGravity = 5
	AttemptsPostAuth       = 3
	AttemptsDeleteSession  = 3
)

func withRetry(retryFunc func() error, attempts uint) error {
	return retry.Do(
		func() error {
			return retryFunc()
		},
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.OnRetry(func(n uint, err error) {
			log.Debug().Msg(fmt.Sprintf("Retrying(%d): %v", n+1, err))
		}),
	)
}
