package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadEnvFile(t *testing.T) {
	os.Clearenv()
	err := LoadEnvFile("../../testdata/.env")

	require.NoError(t, err)

	assert.Equal(t, "https://ph1.example.com|password", os.Getenv("PRIMARY"))
	assert.Equal(t, "https://ph2.example.com|password", os.Getenv("REPLICAS"))
	assert.Equal(t, "false", os.Getenv("FULL_SYNC"))
	assert.Equal(t, "* * * * *", os.Getenv("CRON"))
	assert.Equal(t, "Europe/London", os.Getenv("TZ"))

	assert.Equal(t, "true", os.Getenv("CLIENT_SKIP_TLS_VERIFICATION"))
	assert.Equal(t, "40", os.Getenv("CLIENT_TIMEOUT_SECONDS"))
	assert.Equal(t, "5", os.Getenv("CLIENT_RETRY_DELAY_SECONDS"))

	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_DNS"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_DHCP"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_NTP"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_RESOLVER"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_DATABASE"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_MISC"))
	assert.Equal(t, "true", os.Getenv("SYNC_CONFIG_DEBUG"))

	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_DHCP_LEASES"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_GROUP"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_AD_LIST"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_AD_LIST_BY_GROUP"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_DOMAIN_LIST"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_DOMAIN_LIST_BY_GROUP"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_CLIENT"))
	assert.Equal(t, "true", os.Getenv("SYNC_GRAVITY_CLIENT_BY_GROUP"))

	os.Clearenv()
}

func TestConfig_LoadEnvFile_precedence(t *testing.T) {
	assert.Empty(t, os.Getenv("CRON"))
	t.Setenv("CRON", "0 0 * * *")

	err := LoadEnvFile("../../testdata/.env")
	require.NoError(t, err)

	assert.Equal(t, "0 0 * * *", os.Getenv("CRON"))

	os.Clearenv()
}
