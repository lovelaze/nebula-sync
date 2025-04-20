package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
)

func TestTarget_FullSync(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := NewTarget(primary, []pihole.Client{replica})

	primary.EXPECT().PostAuth().Once().Return(nil)
	replica.EXPECT().PostAuth().Once().Return(nil)

	primary.EXPECT().GetTeleporter().Once().Return([]byte{}, nil)
	replica.EXPECT().PostTeleporter(mock.Anything, mock.Anything).Once().Return(nil)

	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	replica.EXPECT().PatchConfig(mock.Anything).Once().Return(nil)

	primary.EXPECT().PostRunGravity().Once().Return(nil)
	replica.EXPECT().PostRunGravity().Once().Return(nil)

	primary.EXPECT().DeleteSession().Once().Return(nil)
	replica.EXPECT().DeleteSession().Once().Return(nil)

	err := target.FullSync(&config.Sync{
		FullSync:   true,
		Cron:       nil,
		RunGravity: true,
	})
	require.NoError(t, err)
}

func Test_newFullSyncConfigSettings(t *testing.T) {
	gravitySettings := newFullSyncGravitySettings()

	assert.True(t, gravitySettings.DHCPLeases)
	assert.True(t, gravitySettings.Adlist)
	assert.True(t, gravitySettings.AdlistByGroup)
	assert.True(t, gravitySettings.Domainlist)
	assert.True(t, gravitySettings.DomainlistByGroup)
	assert.True(t, gravitySettings.Client)
	assert.True(t, gravitySettings.ClientByGroup)
}

func Test_newFullSyncGravitySettings(t *testing.T) {
	configSettings := newFullSyncConfigSettings()

	assert.True(t, configSettings.DNS.Enabled)
	assert.True(t, configSettings.DHCP.Enabled)
	assert.True(t, configSettings.NTP.Enabled)
	assert.True(t, configSettings.Database.Enabled)
	assert.False(t, configSettings.Webserver.Enabled)
	assert.False(t, configSettings.Files.Enabled)
	assert.True(t, configSettings.Misc.Enabled)
	assert.True(t, configSettings.Debug.Enabled)
}
