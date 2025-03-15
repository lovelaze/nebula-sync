package retry

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/config"
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

var (
	delay time.Duration
)

func Init(clientConfig *config.Client) {
	delay = time.Duration(clientConfig.RetryDelay) * time.Second
}

func Fixed(retryFunc func() error, attempts uint) error {
	return retry.Do(
		func() error {
			return retryFunc()
		},
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Debug().Msg(fmt.Sprintf("Retrying(%d): %v", n+1, err))
		}),
	)
}
