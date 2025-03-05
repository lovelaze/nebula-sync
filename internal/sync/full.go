package sync

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/rs/zerolog/log"
)

func (target *target) FullSync(syncConf *config.Sync) error {
	log.Info().Str("mode", "full").Int("replicas", len(target.Replicas)).Msg("Running sync")
	gravitySettings := newFullSyncGravitySettings()
	configSettings := newFullSyncConfigSettings()

	if err := target.authenticate(); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if err := target.syncTeleporters(gravitySettings); err != nil {
		return fmt.Errorf("sync teleporters: %w", err)
	}

	if err := target.syncConfigs(configSettings); err != nil {
		return fmt.Errorf("sync configs: %w", err)
	}

	if syncConf.RunGravity {
		if err := target.runGravity(); err != nil {
			return fmt.Errorf("run gravity: %w", err)
		}
	}

	if err := target.deleteSessions(); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	return nil
}

func newFullSyncConfigSettings() *config.ConfigSettings {
	return &config.ConfigSettings{
		DNS:       true,
		DHCP:      true,
		NTP:       true,
		Resolver:  true,
		Database:  true,
		Webserver: false,
		Files:     false,
		Misc:      true,
		Debug:     true,
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
