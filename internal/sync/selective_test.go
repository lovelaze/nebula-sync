package sync

import (
	"testing"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTarget_SelectiveSync(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := NewTarget(primary, []pihole.Client{replica}, mockClient)

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
			DNS:       true,
			DHCP:      true,
			NTP:       true,
			Resolver:  true,
			Database:  true,
			Webserver: false,
			Files:     false,
			Misc:      true,
			Debug:     true,
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
