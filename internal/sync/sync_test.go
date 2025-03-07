package sync

import (
	"testing"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/assert"
)

func Test_target_authenticate(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	primary.EXPECT().PostAuth().Once().Return(nil)
	replica.EXPECT().PostAuth().Once().Return(nil)

	err := target.authenticate()
	assert.NoError(t, err)
}

func Test_target_deleteSessions(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	primary.EXPECT().DeleteSession().Once().Return(nil)
	replica.EXPECT().DeleteSession().Once().Return(nil)

	target.deleteSessions()
}

func Test_target_syncTeleporters(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	gravitySettings := config.GravitySettings{
		DHCPLeases:        false,
		Group:             false,
		Adlist:            false,
		AdlistByGroup:     false,
		Domainlist:        false,
		DomainlistByGroup: false,
		Client:            false,
		ClientByGroup:     false,
	}

	primary.EXPECT().GetTeleporter().Once().Return([]byte{}, nil)
	replica.EXPECT().PostTeleporter([]byte{}, createPostTeleporterRequest(&gravitySettings)).Once().Return(nil)

	err := target.syncTeleporters(&gravitySettings)
	assert.NoError(t, err)
}

func Test_target_syncConfigs(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	configResponse := model.ConfigResponse{Config: make(map[string]interface{})}

	gravitySettings := config.ConfigSettings{
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

	primary.EXPECT().GetConfig().Once().Return(&configResponse, nil)
	replica.EXPECT().PatchConfig(createPatchConfigRequest(&gravitySettings, &configResponse)).Once().Return(nil)

	err := target.syncConfigs(&gravitySettings)
	assert.NoError(t, err)
}

func Test_target_runGravity(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipSSLVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	primary.EXPECT().PostRunGravity().Once().Return(nil)
	replica.EXPECT().PostRunGravity().Once().Return(nil)

	err := target.runGravity()
	assert.NoError(t, err)
}

func emptyConfigResponse() *model.ConfigResponse {
	return &model.ConfigResponse{Config: map[string]interface{}{
		"dns":      map[string]interface{}{},
		"dhcp":     map[string]interface{}{},
		"ntp":      map[string]interface{}{},
		"resolver": map[string]interface{}{},
		"database": map[string]interface{}{},
		"misc":     map[string]interface{}{},
		"debug":    map[string]interface{}{},
	}}
}
