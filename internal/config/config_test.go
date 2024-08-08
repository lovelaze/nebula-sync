package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	conf := Config{}

	t.Setenv("PRIMARY", "http://localhost:1337|asdf")
	t.Setenv("REPLICAS", "http://localhost:1338|qwerty")
	t.Setenv("FULL_SYNC", "true")
	t.Setenv("CRON", "* * * * *")

	conf.Load()

	assert.Equal(t, "http://localhost:1337", conf.Primary.Url.String())
	assert.Equal(t, "asdf", conf.Primary.Password)
	assert.Len(t, conf.Replicas, 1)
	assert.Equal(t, "http://localhost:1338", conf.Replicas[0].Url.String())
	assert.Equal(t, "qwerty", conf.Replicas[0].Password)
	assert.Equal(t, true, conf.FullSync)
	assert.Equal(t, "* * * * *", *conf.Cron)
	assert.Nil(t, conf.SyncSettings)
}

func TestConfig_loadSyncSettings(t *testing.T) {
	conf := Config{}
	assert.Nil(t, conf.SyncSettings)

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

	conf.loadSyncSettings()

	assert.NotNil(t, conf.SyncSettings.Config)
	assert.NotNil(t, conf.SyncSettings.Gravity)

	assert.True(t, conf.SyncSettings.Config.DNS)
	assert.True(t, conf.SyncSettings.Config.DHCP)
	assert.True(t, conf.SyncSettings.Config.NTP)
	assert.True(t, conf.SyncSettings.Config.Resolver)
	assert.True(t, conf.SyncSettings.Config.Database)
	assert.True(t, conf.SyncSettings.Config.Misc)
	assert.True(t, conf.SyncSettings.Config.Debug)

	assert.True(t, conf.SyncSettings.Gravity.DHCPLeases)
	assert.True(t, conf.SyncSettings.Gravity.Group)
	assert.True(t, conf.SyncSettings.Gravity.Adlist)
	assert.True(t, conf.SyncSettings.Gravity.AdlistByGroup)
	assert.True(t, conf.SyncSettings.Gravity.Domainlist)
	assert.True(t, conf.SyncSettings.Gravity.DomainlistByGroup)
	assert.True(t, conf.SyncSettings.Gravity.Client)
	assert.True(t, conf.SyncSettings.Gravity.ClientByGroup)
}
