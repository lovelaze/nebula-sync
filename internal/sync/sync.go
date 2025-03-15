package sync

import (
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/retry"
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

func (target *target) authenticate() (err error) {
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

	return err
}

func (target *target) deleteSessions() {
	log.Info().Msg("Invalidating sessions...")
	if err := target.Primary.DeleteSession(); err != nil {
		log.Warn().Msgf("Failed to close session for target : %s", target.Primary.String())
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.DeleteSession()
		}, retry.AttemptsDeleteSession); err != nil {
			log.Warn().Msgf("Failed to close session for target : %s", replica.String())
		}
	}
}

func (target *target) syncTeleporters(gravitySettings *config.GravitySettings) error {
	log.Info().Msg("Syncing Teleporters...")
	conf, err := target.Primary.GetTeleporter()
	if err != nil {
		return err
	}

	var teleporterRequest *model.PostTeleporterRequest = nil
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

	err := target.Primary.PostRunGravity()
	if err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		if err := retry.Fixed(func() error {
			return replica.PostRunGravity()
		}, retry.AttemptsPostRunGravity); err != nil {
			return err
		}
	}

	return err
}

func createPatchConfigRequest(config *config.ConfigSettings, configResponse *model.ConfigResponse) *model.PatchConfigRequest {
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
