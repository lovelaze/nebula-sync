package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	conf := Config{}

	t.Setenv("PRIMARY", "http://localhost:1337|asdf")
	t.Setenv("REPLICAS", "http://localhost:1338|qwerty")
	t.Setenv("FULL_SYNC", "false")

	err := conf.Load()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:1337", conf.Primary.Url.String())
	assert.Equal(t, "asdf", conf.Primary.Password)
	assert.Len(t, conf.Replicas, 1)
	assert.Equal(t, "http://localhost:1338", conf.Replicas[0].Url.String())
	assert.Equal(t, "qwerty", conf.Replicas[0].Password)
	assert.Equal(t, false, conf.Sync.FullSync)
}

func TestConfig_loadSync(t *testing.T) {
	conf := Config{}
	assert.Nil(t, conf.Sync)

	t.Setenv("FULL_SYNC", "true")
	t.Setenv("CRON", "* * * * *")
	t.Setenv("RUN_GRAVITY", "true")

	t.Setenv("SYNC_CONFIG_DNS", "true")
	t.Setenv("SYNC_CONFIG_DHCP", "true")
	t.Setenv("SYNC_CONFIG_NTP", "true")
	t.Setenv("SYNC_CONFIG_RESOLVER", "true")
	t.Setenv("SYNC_CONFIG_DATABASE", "true")
	t.Setenv("SYNC_CONFIG_MISC", "true")
	t.Setenv("SYNC_CONFIG_DEBUG", "true")

	t.Setenv("SYNC_GRAVITY_DHCP_LEASES", "true")
	t.Setenv("SYNC_GRAVITY_GROUP", "true")
	t.Setenv("SYNC_GRAVITY_AD_LIST", "true")
	t.Setenv("SYNC_GRAVITY_AD_LIST_BY_GROUP", "true")
	t.Setenv("SYNC_GRAVITY_DOMAIN_LIST", "true")
	t.Setenv("SYNC_GRAVITY_DOMAIN_LIST_BY_GROUP", "true")
	t.Setenv("SYNC_GRAVITY_CLIENT", "true")
	t.Setenv("SYNC_GRAVITY_CLIENT_BY_GROUP", "true")

	err := conf.loadSync()
	require.NoError(t, err)

	assert.Equal(t, true, conf.Sync.FullSync)
	assert.Equal(t, "* * * * *", *conf.Sync.Cron)
	assert.Equal(t, true, conf.Sync.RunGravity)

	assert.NotNil(t, conf.Sync.ConfigSettings)
	assert.NotNil(t, conf.Sync.GravitySettings)

	assert.True(t, conf.Sync.ConfigSettings.DNS)
	assert.True(t, conf.Sync.ConfigSettings.DHCP)
	assert.True(t, conf.Sync.ConfigSettings.NTP)
	assert.True(t, conf.Sync.ConfigSettings.Resolver)
	assert.True(t, conf.Sync.ConfigSettings.Database)
	assert.True(t, conf.Sync.ConfigSettings.Misc)
	assert.True(t, conf.Sync.ConfigSettings.Debug)

	assert.True(t, conf.Sync.GravitySettings.DHCPLeases)
	assert.True(t, conf.Sync.GravitySettings.Group)
	assert.True(t, conf.Sync.GravitySettings.Adlist)
	assert.True(t, conf.Sync.GravitySettings.AdlistByGroup)
	assert.True(t, conf.Sync.GravitySettings.Domainlist)
	assert.True(t, conf.Sync.GravitySettings.DomainlistByGroup)
	assert.True(t, conf.Sync.GravitySettings.Client)
	assert.True(t, conf.Sync.GravitySettings.ClientByGroup)
}

func TestConfig_LoadClient(t *testing.T) {
	conf := Config{}

	t.Setenv("CLIENT_SKIP_TLS_VERIFICATION", "true")
	t.Setenv("CLIENT_TIMEOUT_SECONDS", "45")

	err := conf.loadClient()
	require.NoError(t, err)

	assert.Equal(t, true, conf.Client.SkipSSLVerification)
	assert.Equal(t, uint(45), conf.Client.Timeout)
}

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
}
