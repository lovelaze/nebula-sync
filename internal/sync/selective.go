package sync

import (
	"fmt"

	"github.com/lovelaze/nebula-sync/internal/config"
)

func (target *target) SelectiveSync(conf *config.Sync) error {
	return target.sync(func() error {
		return target.selective(conf)
	}, "selective")
}

func (target *target) selective(conf *config.Sync) error {
	if err := target.syncTeleporters(conf.GravitySettings); err != nil {
		return fmt.Errorf("sync teleporters: %w", err)
	}

	if err := target.syncConfigs(conf.ConfigSettings, conf.ExcludeDNSRecords); err != nil {
		return fmt.Errorf("sync configs: %w", err)
	}

	if conf.RunGravity {
		if err := target.runGravity(); err != nil {
			return fmt.Errorf("run gravity: %w", err)
		}
	}
	return nil
}
