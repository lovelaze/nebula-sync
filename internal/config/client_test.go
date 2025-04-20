package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadClient(t *testing.T) {
	conf := Config{}

	t.Setenv("CLIENT_SKIP_TLS_VERIFICATION", "true")
	t.Setenv("CLIENT_TIMEOUT_SECONDS", "45")
	t.Setenv("CLIENT_RETRY_DELAY_SECONDS", "5")

	err := conf.loadClient()
	require.NoError(t, err)

	assert.True(t, conf.Client.SkipTLSVerification)
	assert.Equal(t, int64(45), conf.Client.Timeout)
	assert.Equal(t, int64(5), conf.Client.RetryDelay)
}
