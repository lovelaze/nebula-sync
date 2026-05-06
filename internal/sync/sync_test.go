package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
)

func Test_target_authenticate(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipTLSVerification: false,
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
		SkipTLSVerification: false,
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
		SkipTLSVerification: false,
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
		SkipTLSVerification: false,
		RetryDelay:          1,
	}

	target := target{
		Primary:  primary,
		Replicas: []pihole.Client{replica},
		Client:   mockClient,
	}

	configResponse := emptyConfigResponse()

	gravitySettings := config.ConfigSettings{
		DNS:       config.NewConfigSetting(false, nil, nil),
		DHCP:      config.NewConfigSetting(false, nil, nil),
		NTP:       config.NewConfigSetting(false, nil, nil),
		Resolver:  config.NewConfigSetting(false, nil, nil),
		Database:  config.NewConfigSetting(false, nil, nil),
		Webserver: config.NewConfigSetting(false, nil, nil),
		Files:     config.NewConfigSetting(false, nil, nil),
		Misc:      config.NewConfigSetting(false, nil, nil),
		Debug:     config.NewConfigSetting(false, nil, nil),
	}

	primary.EXPECT().GetConfig().Once().Return(configResponse, nil)
	replica.EXPECT().PatchConfig(createPatchConfigRequest(&gravitySettings, configResponse, nil, nil)).Once().Return(nil)

	err := target.syncConfigs(&gravitySettings, nil)
	assert.NoError(t, err)
}

func Test_target_runGravity(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)

	mockClient := &config.Client{
		SkipTLSVerification: false,
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

func Test_filterPatchConfigRequest_enabled(t *testing.T) {
	dns := emptyConfigResponse().Get("dns")

	request := filterPatchConfigRequest(&config.ConfigSetting{
		Enabled: true,
		Filter:  nil,
	}, dns)
	assert.Equal(t, dns, request)
}

func Test_filterPatchConfigRequest_disabled(t *testing.T) {
	dns := emptyConfigResponse().Get("dns")

	request := filterPatchConfigRequest(&config.ConfigSetting{
		Enabled: false,
		Filter:  nil,
	}, dns)
	assert.Nil(t, request)
}

func Test_mergeDNSRecords(t *testing.T) {
	primaryDNS := map[string]any{
		"hosts": []any{
			"192.168.1.1 example.com",
			"192.168.1.2 test.com",
			"192.168.1.3 other.com",
		},
		"cnameRecords": []any{
			"alias.example.com,example.com",
			"alias.test.com,test.com",
		},
	}

	replicaDNS := map[string]any{
		"hosts": []any{
			"192.168.1.100 example.com",
			"192.168.1.101 test.com",
			"192.168.1.102 replica.com",
		},
		"cnameRecords": []any{
			"replica-alias.test.com,test.com",
		},
	}

	excludeList := []string{"example.com", "test.com"}

	merged := mergeDNSRecords(primaryDNS, replicaDNS, excludeList)

	// Should have 1 from primary (other.com) and 2 from replica (example.com, test.com)
	assert.Len(t, merged["hosts"].([]any), 3)
	assert.Contains(t, merged["hosts"].([]any), "192.168.1.3 other.com")
	assert.Contains(t, merged["hosts"].([]any), "192.168.1.100 example.com")
	assert.Contains(t, merged["hosts"].([]any), "192.168.1.101 test.com")
	assert.NotContains(t, merged["hosts"].([]any), "192.168.1.102 replica.com")

	// Should have 0 from primary and 1 from replica
	assert.Len(t, merged["cnameRecords"].([]any), 1)
	assert.Equal(t, "replica-alias.test.com,test.com", merged["cnameRecords"].([]any)[0])
}

func emptyConfigResponse() *model.ConfigResponse {
	return &model.ConfigResponse{Config: map[string]any{
		"dns":      map[string]any{},
		"dhcp":     map[string]any{},
		"ntp":      map[string]any{},
		"resolver": map[string]any{},
		"database": map[string]any{},
		"misc":     map[string]any{},
		"debug":    map[string]any{},
	}}
}
