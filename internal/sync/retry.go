package sync

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/rs/zerolog/log"
)

const (
	AttemptsPostTeleporter = 5
	AttemptsPatchConfig    = 5
	AttemptsPostRunGravity = 5
	AttemptsPostAuth       = 3
	AttemptsDeleteSession  = 3
)

func withRetry(retryFunc func() error, attempts, delay uint) error {
	return retry.Do(
		func() error {
			return retryFunc()
		},
		retry.Attempts(attempts),
		retry.Delay(time.Duration(delay)*time.Second),
		retry.OnRetry(func(n uint, err error) {
			log.Debug().Msg(fmt.Sprintf("Retrying(%d): %v", n+1, err))
		}),
	)
}
