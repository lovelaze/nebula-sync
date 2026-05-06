package sync

import (
	"fmt"

	"github.com/lovelaze/nebula-sync/internal/config"
)

func (target *target) FullSync(conf *config.Sync) error {
	return target.sync(func() error {
		return target.full(conf)
	}, "full")
}

func (target *target) full(conf *config.Sync) error {
	gravitySettings := newFullSyncGravitySettings()
	configSettings := newFullSyncConfigSettings()

	if err := target.syncTeleporters(gravitySettings); err != nil {
		return fmt.Errorf("sync teleporters: %w", err)
	}

	if err := target.syncConfigs(configSettings, conf.ExcludeDNSRecords); err != nil {
		return fmt.Errorf("sync configs: %w", err)
	}

	if conf.RunGravity {
		if err := target.runGravity(); err != nil {
			return fmt.Errorf("run gravity: %w", err)
		}
	}
	return nil
}

func newFullSyncConfigSettings() *config.ConfigSettings {
	return &config.ConfigSettings{
		DNS:       config.NewConfigSetting(true, nil, nil),
		DHCP:      config.NewConfigSetting(true, nil, nil),
		NTP:       config.NewConfigSetting(true, nil, nil),
		Resolver:  config.NewConfigSetting(true, nil, nil),
		Database:  config.NewConfigSetting(true, nil, nil),
		Webserver: config.NewConfigSetting(false, nil, nil),
		Files:     config.NewConfigSetting(false, nil, nil),
		Misc:      config.NewConfigSetting(true, nil, nil),
		Debug:     config.NewConfigSetting(true, nil, nil),
	}
}

func newFullSyncGravitySettings() *config.GravitySettings {
	return &config.GravitySettings{
		DHCPLeases:        true,
		Group:             true,
		Adlist:            true,
		AdlistByGroup:     true,
		Domainlist:        true,
		DomainlistByGroup: true,
		Client:            true,
		ClientByGroup:     true,
	}
}
