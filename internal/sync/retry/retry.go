package retry

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/config"
)

const (
	AttemptsPostTeleporter = 5
	AttemptsPatchConfig    = 5
	AttemptsPostRunGravity = 5
	AttemptsPostAuth       = 3
	AttemptsDeleteSession  = 3
)

var delay time.Duration

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
		retry.LastErrorOnly(true),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Debug().Msg(fmt.Sprintf("Retrying(%d): %v", n+1, err))
		}),
	)
}
