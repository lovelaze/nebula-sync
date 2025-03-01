package sync

import (
	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTarget_FullSync(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := NewTarget(primary, []pihole.Client{replica})

	primary.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)

	primary.
		EXPECT().
		GetTeleporter().
		Times(1).
		Return([]byte{}, nil)
	replica.
		EXPECT().
		PostTeleporter(mock.Anything, mock.Anything).
		Times(1).
		Return(nil)

	primary.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)

	err := target.FullSync()
	require.NoError(t, err)
}

func TestTarget_ManualSync(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := NewTarget(primary, []pihole.Client{replica})

	settings := config.SyncSettings{
		Gravity: &config.ManualGravity{
			DHCPLeases:        false,
			Group:             false,
			Adlist:            false,
			AdlistByGroup:     false,
			Domainlist:        false,
			DomainlistByGroup: false,
			Client:            false,
			ClientByGroup:     false,
		},
		Config: &config.ManualConfig{
			DNS:       false,
			DHCP:      false,
			NTP:       false,
			Resolver:  false,
			Database:  false,
			Webserver: false,
			Files:     false,
			Misc:      false,
			Debug:     false,
		},
	}

	primary.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)

	primary.
		EXPECT().
		GetTeleporter().
		Times(1).
		Return([]byte{}, nil)
	replica.
		EXPECT().
		PostTeleporter(mock.Anything, mock.Anything).
		Times(1).
		Return(nil)

	primary.
		EXPECT().
		GetConfig().
		Times(1).
		Return(&model.ConfigResponse{Config: make(map[string]interface{})}, nil)
	replica.
		EXPECT().
		PatchConfig(mock.Anything).
		Times(1).
		Return(nil)

	primary.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)

	err := target.ManualSync(&settings)
	require.NoError(t, err)
}

func Test_target_authenticate(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
	}

	primary.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		PostAuth().
		Times(1).
		Return(nil)

	err := target.authenticate()
	assert.NoError(t, err)
}

func Test_target_deleteSessions(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
	}

	primary.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)
	replica.
		EXPECT().
		DeleteSession().
		Times(1).
		Return(nil)

	err := target.deleteSessions()
	assert.NoError(t, err)
}

func Test_target_syncTeleporters(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
	}

	manualGravity := config.ManualGravity{
		DHCPLeases:        false,
		Group:             false,
		Adlist:            false,
		AdlistByGroup:     false,
		Domainlist:        false,
		DomainlistByGroup: false,
		Client:            false,
		ClientByGroup:     false,
	}

	primary.
		EXPECT().
		GetTeleporter().
		Times(1).
		Return([]byte{}, nil)
	replica.
		EXPECT().
		PostTeleporter([]byte{}, createPostTeleporterRequest(&manualGravity)).
		Times(1).
		Return(nil)

	err := target.syncTeleporters(&manualGravity)
	assert.NoError(t, err)
}

func Test_target_syncConfigs(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
	}

	configResponse := model.ConfigResponse{Config: make(map[string]interface{})}

	manualConfig := config.ManualConfig{
		DNS:       false,
		DHCP:      false,
		NTP:       false,
		Resolver:  false,
		Database:  false,
		Webserver: false,
		Files:     false,
		Misc:      false,
		Debug:     false,
	}

	primary.
		EXPECT().
		GetConfig().
		Times(1).
		Return(&configResponse, nil)
	replica.
		EXPECT().
		PatchConfig(createPatchConfigRequest(&manualConfig, &configResponse)).
		Times(1).
		Return(nil)

	err := target.syncConfigs(&manualConfig)
	assert.NoError(t, err)
}
