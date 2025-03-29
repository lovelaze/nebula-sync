package config

import (
	"github.com/lovelaze/nebula-sync/internal/sync/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.True(t, conf.Sync.ConfigSettings.DNS.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.DHCP.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.NTP.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.Resolver.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.Database.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.Misc.Enabled)
	assert.True(t, conf.Sync.ConfigSettings.Debug.Enabled)

	assert.True(t, conf.Sync.GravitySettings.DHCPLeases)
	assert.True(t, conf.Sync.GravitySettings.Group)
	assert.True(t, conf.Sync.GravitySettings.Adlist)
	assert.True(t, conf.Sync.GravitySettings.AdlistByGroup)
	assert.True(t, conf.Sync.GravitySettings.Domainlist)
	assert.True(t, conf.Sync.GravitySettings.DomainlistByGroup)
	assert.True(t, conf.Sync.GravitySettings.Client)
	assert.True(t, conf.Sync.GravitySettings.ClientByGroup)
}

func TestRawConfig_Validate_Both(t *testing.T) {
	settings := RawConfigSettings{
		DNSInclude: []string{"a"},
		DNSExclude: []string{"b"},
	}
	assert.Error(t, settings.Validate())
}

func TestRawConfig_Validate_Single(t *testing.T) {
	include := RawConfigSettings{
		DNSInclude: []string{"a"},
		DNSExclude: nil,
	}
	exclude := RawConfigSettings{
		DNSInclude: nil,
		DNSExclude: []string{"a"},
	}
	assert.NoError(t, include.Validate())
	assert.NoError(t, exclude.Validate())
}

func TestRawConfig_Validate_None(t *testing.T) {
	settings := RawConfigSettings{
		DNSInclude: nil,
		DNSExclude: nil,
	}
	assert.NoError(t, settings.Validate())
}

func TestRawConfig_Parse_Include(t *testing.T) {
	t.Setenv("SYNC_CONFIG_DNS_INCLUDE", "key1,key2")
	t.Setenv("SYNC_CONFIG_DHCP_INCLUDE", "key3,key4")
	t.Setenv("SYNC_CONFIG_NTP_INCLUDE", "key5,key6")
	t.Setenv("SYNC_CONFIG_RESOLVER_INCLUDE", "key7,key8")
	t.Setenv("SYNC_CONFIG_DATABASE_INCLUDE", "key9,key10")
	t.Setenv("SYNC_CONFIG_MISC_INCLUDE", "key11,key12")
	t.Setenv("SYNC_CONFIG_DEBUG_INCLUDE", "key13,key14")

	sync := Sync{}
	assert.NoError(t, sync.loadConfigSettings())

	settings := sync.ConfigSettings

	assert.Equal(t, settings.DNS.Filter.Type, filter.Include)
	assert.Equal(t, settings.DNS.Filter.Keys, []string{"key1", "key2"})
	assert.Equal(t, settings.DHCP.Filter.Type, filter.Include)
	assert.Equal(t, settings.DHCP.Filter.Keys, []string{"key3", "key4"})
	assert.Equal(t, settings.NTP.Filter.Type, filter.Include)
	assert.Equal(t, settings.NTP.Filter.Keys, []string{"key5", "key6"})
	assert.Equal(t, settings.Resolver.Filter.Type, filter.Include)
	assert.Equal(t, settings.Resolver.Filter.Keys, []string{"key7", "key8"})
	assert.Equal(t, settings.Database.Filter.Type, filter.Include)
	assert.Equal(t, settings.Database.Filter.Keys, []string{"key9", "key10"})
	assert.Equal(t, settings.Misc.Filter.Type, filter.Include)
	assert.Equal(t, settings.Misc.Filter.Keys, []string{"key11", "key12"})
	assert.Equal(t, settings.Debug.Filter.Type, filter.Include)
	assert.Equal(t, settings.Debug.Filter.Keys, []string{"key13", "key14"})
}

func TestRawConfig_Parse_Exclude(t *testing.T) {
	t.Setenv("SYNC_CONFIG_DNS_EXCLUDE", "key1,key2")
	t.Setenv("SYNC_CONFIG_DHCP_EXCLUDE", "key3,key4")
	t.Setenv("SYNC_CONFIG_NTP_EXCLUDE", "key5,key6")
	t.Setenv("SYNC_CONFIG_RESOLVER_EXCLUDE", "key7,key8")
	t.Setenv("SYNC_CONFIG_DATABASE_EXCLUDE", "key9,key10")
	t.Setenv("SYNC_CONFIG_MISC_EXCLUDE", "key11,key12")
	t.Setenv("SYNC_CONFIG_DEBUG_EXCLUDE", "key13,key14")

	sync := Sync{}
	assert.NoError(t, sync.loadConfigSettings())

	settings := sync.ConfigSettings

	assert.Equal(t, settings.DNS.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.DNS.Filter.Keys, []string{"key1", "key2"})
	assert.Equal(t, settings.DHCP.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.DHCP.Filter.Keys, []string{"key3", "key4"})
	assert.Equal(t, settings.NTP.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.NTP.Filter.Keys, []string{"key5", "key6"})
	assert.Equal(t, settings.Resolver.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.Resolver.Filter.Keys, []string{"key7", "key8"})
	assert.Equal(t, settings.Database.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.Database.Filter.Keys, []string{"key9", "key10"})
	assert.Equal(t, settings.Misc.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.Misc.Filter.Keys, []string{"key11", "key12"})
	assert.Equal(t, settings.Debug.Filter.Type, filter.Exclude)
	assert.Equal(t, settings.Debug.Filter.Keys, []string{"key13", "key14"})
}

func TestConfig_NewConfigSetting(t *testing.T) {
	enabled := NewConfigSetting(true, nil, nil)
	assert.True(t, enabled.Enabled)
	assert.Nil(t, enabled.Filter)

	disabled := NewConfigSetting(false, nil, nil)
	assert.False(t, disabled.Enabled)
	assert.Nil(t, disabled.Filter)

	include := NewConfigSetting(true, []string{"key1", "key2"}, nil)
	assert.True(t, include.Enabled)
	assert.NotNil(t, include.Filter)
	assert.Equal(t, include.Filter.Type, filter.Include)
	assert.Equal(t, include.Filter.Keys, []string{"key1", "key2"})

	exclude := NewConfigSetting(true, nil, []string{"key1", "key2"})
	assert.True(t, exclude.Enabled)
	assert.NotNil(t, exclude.Filter)
	assert.Equal(t, exclude.Filter.Type, filter.Exclude)
	assert.Equal(t, exclude.Filter.Keys, []string{"key1", "key2"})
}
