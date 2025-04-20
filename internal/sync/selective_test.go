package sync

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
)

func TestTarget_SelectiveSync(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := NewTarget(primary, []pihole.Client{replica})

	settings := config.Sync{
		FullSync:   false,
		RunGravity: true,
		GravitySettings: &config.GravitySettings{
			DHCPLeases:        true,
			Group:             true,
			Adlist:            true,
			AdlistByGroup:     true,
			Domainlist:        true,
			DomainlistByGroup: true,
			Client:            true,
			ClientByGroup:     true,
		},
		ConfigSettings: &config.ConfigSettings{
			DNS:       config.NewConfigSetting(true, nil, nil),
			DHCP:      config.NewConfigSetting(true, nil, nil),
			NTP:       config.NewConfigSetting(true, nil, nil),
			Resolver:  config.NewConfigSetting(true, nil, nil),
			Database:  config.NewConfigSetting(true, nil, nil),
			Webserver: config.NewConfigSetting(false, nil, nil),
			Files:     config.NewConfigSetting(false, nil, nil),
			Misc:      config.NewConfigSetting(true, nil, nil),
			Debug:     config.NewConfigSetting(true, nil, nil),
		},
	}

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

	err := target.SelectiveSync(&settings)
	require.NoError(t, err)
}
