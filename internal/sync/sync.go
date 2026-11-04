package sync

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/sync/filter"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
)

type Target interface {
	FullSync(sync *config.Sync) error
	SelectiveSync(sync *config.Sync) error
}

type target struct {
	Primary  pihole.Client
	Replicas []pihole.Client
	Client   *config.Client
	changes  *changeTracker
}

func NewTarget(primary pihole.Client, replicas []pihole.Client) Target {
	return &target{
		Primary:  primary,
		Replicas: replicas,
		changes:  newChangeTracker(),
	}
}

func (target *target) sync(
	syncFunc func() error,
	mode string,
	conf *config.Sync,
	gravitySettings *config.GravitySettings,
	configSettings *config.ConfigSettings,
) error {
	var err error
	replicasContacted := false
	log.Info().Str("mode", mode).Int("replicas", len(target.Replicas)).Msg("Running sync")

	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Error during sync")
		}
		log.Info().Msg("Invalidating sessions...")
		target.deletePrimarySession()
		if replicasContacted {
			target.deleteReplicaSessions()
		}
	}()

	log.Info().Msg("Authenticating clients...")
	if err := target.authenticatePrimary(); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	// With SYNC_ON_CHANGE the primary is checked before the replicas are contacted,
	// so a skipped run opens no replica sessions and does not fail on a replica
	// that is down.
	var fingerprint string
	if conf != nil && conf.SyncOnChange {
		var skip bool
		fingerprint, skip = target.checkChanges(gravitySettings, configSettings, conf.SyncOnChangeMaxAge)
		if skip {
			return ErrNoChanges
		}
	}

	replicasContacted = true
	if err := target.authenticateReplicas(); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if err := syncFunc(); err != nil {
		return err
	}

	if fingerprint != "" {
		target.changes.record(fingerprint)
	}
	return nil
}

// checkChanges returns the primary's fingerprint and whether the sync can be
// skipped. On error it returns an empty fingerprint and syncs anyway.
func (target *target) checkChanges(
	gravitySettings *config.GravitySettings,
	configSettings *config.ConfigSettings,
	maxAge time.Duration,
) (string, bool) {
	fingerprint, err := target.fingerprint(gravitySettings, configSettings)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to check primary for changes, syncing anyway")
		return "", false
	}

	due, reason := target.changes.due(fingerprint, maxAge)
	if !due {
		log.Info().Msg("No changes on primary since last sync, skipping")
		return fingerprint, true
	}

	log.Info().Str("reason", reason).Msg("Sync on change")
	return fingerprint, false
}

func (target *target) authenticate() error {
	log.Info().Msg("Authenticating clients...")
	if err := target.authenticatePrimary(); err != nil {
		return err
	}
	return target.authenticateReplicas()
}

func (target *target) authenticatePrimary() error {
	return target.Primary.PostAuth()
}

func (target *target) authenticateReplicas() error {
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
	target.deletePrimarySession()
	target.deleteReplicaSessions()
}

func (target *target) deletePrimarySession() {
	if err := target.Primary.DeleteSession(); err != nil {
		log.Warn().Msgf("Failed to invalidate session for target: %s", target.Primary.String())
	}
}

func (target *target) deleteReplicaSessions() {
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
