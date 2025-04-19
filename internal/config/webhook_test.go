package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSettings_Load_Success(t *testing.T) {
	t.Setenv("WEBHOOK_SYNC_SUCCESS_URL", "http://success.example.com")
	t.Setenv("WEBHOOK_SYNC_SUCCESS_METHOD", "POST")
	t.Setenv("WEBHOOK_SYNC_SUCCESS_BODY", "{\"status\":\"ok\"}")
	t.Setenv("WEBHOOK_SYNC_SUCCESS_HEADERS", "Content-Type:application/json,Authorization:Bearer token,X-Custom-Header: CustomValue")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	success := conf.Sync.WebhookSettings.Success
	assert.Equal(t, "http://success.example.com", success.URL)
	assert.Equal(t, "POST", success.Method)
	assert.JSONEq(t, `{"status":"ok"}`, success.Body)
	assert.Equal(t, map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer token",
		"X-Custom-Header": " CustomValue",
	}, success.Headers)
}

func TestWebhookSettings_Load_Failure(t *testing.T) {
	t.Setenv("WEBHOOK_SYNC_FAILURE_URL", "http://failure.example.com")
	t.Setenv("WEBHOOK_SYNC_FAILURE_METHOD", "PUT")
	t.Setenv("WEBHOOK_SYNC_FAILURE_BODY", "{\"status\":\"error\"}")
	t.Setenv("WEBHOOK_SYNC_FAILURE_HEADERS", "Content-Type:application/json")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	failure := conf.Sync.WebhookSettings.Failure
	assert.Equal(t, "http://failure.example.com", failure.URL)
	assert.Equal(t, "PUT", failure.Method)
	assert.JSONEq(t, `{"status":"error"}`, failure.Body)
	assert.Equal(t, map[string]string{
		"Content-Type": "application/json",
	}, failure.Headers)
}

func TestWebhookSettings_DefaultValues(t *testing.T) {
	t.Setenv("WEBHOOK_SYNC_SUCCESS_URL", "http://success.example.com")
	t.Setenv("WEBHOOK_SYNC_FAILURE_URL", "http://failure.example.com")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	// Test default values
	assert.Equal(t, "POST", conf.Sync.WebhookSettings.Success.Method)
	assert.Equal(t, "POST", conf.Sync.WebhookSettings.Failure.Method)
	assert.Empty(t, conf.Sync.WebhookSettings.Success.Body)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.Body)
	assert.Empty(t, conf.Sync.WebhookSettings.Success.Headers)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.Headers)
}

func TestWebhookSettings_InvalidHeaders(t *testing.T) {
	t.Setenv("WEBHOOK_SYNC_SUCCESS_URL", "http://success.example.com")
	t.Setenv("WEBHOOK_SYNC_SUCCESS_HEADERS", "InvalidHeader")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "process webhook env vars")
}

func TestWebhookSettings_EmptyURLs(t *testing.T) {
	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	assert.Empty(t, conf.Sync.WebhookSettings.Success.URL)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.URL)
}

func TestWebhookSettings_ClientConfiguration(t *testing.T) {
	t.Setenv("WEBHOOK_SYNC_SUCCESS_URL", "http://success.example.com")
	t.Setenv("WEBHOOK_CLIENT_SKIP_TLS_VERIFICATION", "true")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	require.NotNil(t, conf.Sync.WebhookSettings.Client)
	assert.True(t, conf.Sync.WebhookSettings.Client.SkipTLSVerification)
}
