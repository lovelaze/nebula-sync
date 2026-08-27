package config

import (
	"crypto/tls"
	"net/http"
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

func TestClient_NewHTTPClient_TLSMinVersion(t *testing.T) {
	client := &Client{SkipTLSVerification: false, Timeout: 20}

	httpClient := client.NewHTTPClient()
	transport, ok := httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	assert.NotNil(t, transport.Proxy)
}
