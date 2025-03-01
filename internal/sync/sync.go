package sync

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/rs/zerolog/log"
)

type Target interface {
	FullSync() error
	ManualSync(syncSettings *config.SyncSettings) error
}

type target struct {
	Primary  pihole.Client
	Replicas []pihole.Client
}

func NewTarget(primary pihole.Client, replicas []pihole.Client) Target {
	return &target{
		Primary:  primary,
		Replicas: replicas,
	}
}

func (target *target) FullSync() error {
	log.Info().Int("replicas", len(target.Replicas)).Msg("Running full sync")

	if err := target.authenticate(); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if err := target.syncTeleporters(nil); err != nil {
		return fmt.Errorf("sync teleporters: %w", err)
	}

	if err := target.deleteSessions(); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	return nil
}

func (target *target) ManualSync(syncSettings *config.SyncSettings) error {
	log.Info().Int("replicas", len(target.Replicas)).Msg("Running manual sync")

	if err := target.authenticate(); err != nil {
		return fmt.Errorf("authentication: %w", err)
	}

	if err := target.syncTeleporters(syncSettings.Gravity); err != nil {
		return fmt.Errorf("sync teleporters: %w", err)
	}

	if err := target.syncConfigs(syncSettings.Config); err != nil {
		return fmt.Errorf("sync configs: %w", err)
	}

	if err := target.deleteSessions(); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	return nil
}

func (target *target) authenticate() (err error) {
	log.Info().Msg("Authenticating clients...")
	if err := target.Primary.PostAuth(); err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		if err := replica.PostAuth(); err != nil {
			return err
		}
	}

	return err
}

func (target *target) deleteSessions() (err error) {
	log.Info().Msg("Invalidating sessions...")
	if err := target.Primary.DeleteSession(); err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		if err := replica.DeleteSession(); err != nil {
			return err
		}
	}

	return err
}

func (target *target) syncTeleporters(manualGravity *config.ManualGravity) error {
	log.Info().Msg("Syncing Teleporters...")
	conf, err := target.Primary.GetTeleporter()
	if err != nil {
		return err
	}

	var teleporterRequest *model.PostTeleporterRequest = nil
	if manualGravity != nil {
		teleporterRequest = createPostTeleporterRequest(manualGravity)
	}

	for _, replica := range target.Replicas {
		if err := withRetry(func() error {
			return replica.PostTeleporter(conf, teleporterRequest)
		}); err != nil {
			return err
		}
	}

	return err
}

func (target *target) syncConfigs(manualConfig *config.ManualConfig) error {
	log.Info().Msg("Syncing configs...")
	configResponse, err := target.Primary.GetConfig()
	if err != nil {
		return err
	}

	configRequest := createPatchConfigRequest(manualConfig, configResponse)

	for _, replica := range target.Replicas {
		if err := withRetry(func() error {
			return replica.PatchConfig(configRequest)
		}); err != nil {
			return err
		}
	}

	return err
}

func createPatchConfigRequest(config *config.ManualConfig, configResponse *model.ConfigResponse) *model.PatchConfigRequest {
	patchConfig := model.PatchConfig{}

	if config.DNS {
		patchConfig.DNS = configResponse.Config["dns"].(map[string]interface{})
	}
	if config.DHCP {
		patchConfig.DHCP = configResponse.Config["dhcp"].(map[string]interface{})
	}
	if config.NTP {
		patchConfig.NTP = configResponse.Config["ntp"].(map[string]interface{})
	}
	if config.Resolver {
		patchConfig.Resolver = configResponse.Config["resolver"].(map[string]interface{})
	}
	if config.Database {
		patchConfig.Database = configResponse.Config["database"].(map[string]interface{})
	}
	if config.Misc {
		patchConfig.Misc = configResponse.Config["misc"].(map[string]interface{})
	}
	if config.Debug {
		patchConfig.Debug = configResponse.Config["debug"].(map[string]interface{})
	}

	return &model.PatchConfigRequest{Config: patchConfig}
}

func createPostTeleporterRequest(gravity *config.ManualGravity) *model.PostTeleporterRequest {
	return &model.PostTeleporterRequest{
		Config:     false,
		DHCPLeases: gravity.DHCPLeases,
		Gravity: model.PostGravityRequest{
			Group:             gravity.Group,
			Adlist:            gravity.Adlist,
			AdlistByGroup:     gravity.AdlistByGroup,
			Domainlist:        gravity.Domainlist,
			DomainlistByGroup: gravity.DomainlistByGroup,
			Client:            gravity.Client,
			ClientByGroup:     gravity.ClientByGroup,
		},
	}
}
