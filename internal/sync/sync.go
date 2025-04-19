package sync

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/sync/filter"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
	"github.com/rs/zerolog/log"
)

type Target interface {
	FullSync(sync *config.Sync) error
	SelectiveSync(sync *config.Sync) error
}

type target struct {
	Primary  pihole.Client
	Replicas []pihole.Client
	Client   *config.Client
}

func NewTarget(primary pihole.Client, replicas []pihole.Client) Target {
	return &target{
		Primary:  primary,
		Replicas: replicas,
	}
}

func (target *target) sync(syncFunc func() error, mode string) error {
	var err error
	log.Info().Str("mode", mode).Int("replicas", len(target.Replicas)).Msg("Running sync")

	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Error during sync")
		}
		target.deleteSessions()
	}()

	if err := target.authenticate(); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	return syncFunc()
}

func (target *target) authenticate() error {
	log.Info().Msg("Authenticating clients...")
	if err := target.Primary.PostAuth(); err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.PostAuth()
		}, retry.AttemptsPostAuth); err != nil {
			return err
		}
	}

	return nil
}

func (target *target) deleteSessions() {
	log.Info().Msg("Invalidating sessions...")
	if err := target.Primary.DeleteSession(); err != nil {
		log.Warn().Msgf("Failed to invalidate session for target: %s", target.Primary.String())
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.DeleteSession()
		}, retry.AttemptsDeleteSession); err != nil {
			log.Warn().Msgf("Failed to invalidate session for target: %s", replica.String())
		}
	}
}

func (target *target) syncTeleporters(gravitySettings *config.GravitySettings) error {
	log.Info().Msg("Syncing teleporters...")
	conf, err := target.Primary.GetTeleporter()
	if err != nil {
		return err
	}

	var teleporterRequest *model.PostTeleporterRequest
	if gravitySettings != nil {
		teleporterRequest = createPostTeleporterRequest(gravitySettings)
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.PostTeleporter(conf, teleporterRequest)
		}, retry.AttemptsPostTeleporter); err != nil {
			return err
		}
	}

	return err
}

func (target *target) syncConfigs(configSettings *config.ConfigSettings) error {
	log.Info().Msg("Syncing configs...")
	configResponse, err := target.Primary.GetConfig()
	if err != nil {
		return err
	}

	configRequest := createPatchConfigRequest(configSettings, configResponse)

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.PatchConfig(configRequest)
		}, retry.AttemptsPatchConfig); err != nil {
			return err
		}
	}

	return err
}

func (target *target) runGravity() error {
	log.Info().Msg("Running gravity...")

	if err := target.Primary.PostRunGravity(); err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.PostRunGravity()
		}, retry.AttemptsPostRunGravity); err != nil {
			return err
		}
	}

	return nil
}

func createPatchConfigRequest(config *config.ConfigSettings, configResponse *model.ConfigResponse) *model.PatchConfigRequest {
	patchConfig := model.PatchConfig{}

	if json := filterPatchConfigRequest(config.DNS, configResponse.Get("dns")); json != nil {
		patchConfig.DNS = json
	}
	if json := filterPatchConfigRequest(config.DHCP, configResponse.Get("dhcp")); json != nil {
		patchConfig.DHCP = json
	}
	if json := filterPatchConfigRequest(config.NTP, configResponse.Get("ntp")); json != nil {
		patchConfig.NTP = json
	}
	if json := filterPatchConfigRequest(config.Resolver, configResponse.Get("resolver")); json != nil {
		patchConfig.Resolver = json
	}
	if json := filterPatchConfigRequest(config.Database, configResponse.Get("database")); json != nil {
		patchConfig.Database = json
	}
	if json := filterPatchConfigRequest(config.Misc, configResponse.Get("misc")); json != nil {
		patchConfig.Misc = json
	}
	if json := filterPatchConfigRequest(config.Debug, configResponse.Get("debug")); json != nil {
		patchConfig.Debug = json
	}

	return &model.PatchConfigRequest{Config: patchConfig}
}

func filterPatchConfigRequest(setting *config.ConfigSetting, json map[string]any) map[string]any {
	if !setting.Enabled {
		return nil
	}

	if setting.Filter != nil {
		filteredJSON, err := filter.ByType(setting.Filter.Type, setting.Filter.Keys, json)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to filter json object")
			return nil
		}
		return filteredJSON
	}

	return json
}

func createPostTeleporterRequest(gravity *config.GravitySettings) *model.PostTeleporterRequest {
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
