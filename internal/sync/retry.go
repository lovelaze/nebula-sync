package sync

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	attempts uint          = 5
	delay    time.Duration = 1 * time.Second
)

func withRetry(retryFunc func() error) error {
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
