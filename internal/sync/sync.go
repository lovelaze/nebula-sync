package sync

import (
	"fmt"
	"strings"

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

func (target *target) syncConfigs(configSettings *config.ConfigSettings, excludeDNSRecords []string) error {
	log.Info().Msg("Syncing configs...")
	primaryResponse, err := target.Primary.GetConfig()
	if err != nil {
		return err
	}

	for _, replica := range target.Replicas {
		var replicaResponse *model.ConfigResponse
		if len(excludeDNSRecords) > 0 {
			var err error
			replicaResponse, err = replica.GetConfig()
			if err != nil {
				log.Warn().Err(err).Msgf(
					"Failed to get config from replica %s, excluded records might be deleted",
					replica.String(),
				)
			}
		}

		configRequest := createPatchConfigRequest(configSettings, primaryResponse, replicaResponse, excludeDNSRecords)

		if err := retry.Fixed(func() error {
			return replica.PatchConfig(configRequest)
		}, retry.AttemptsPatchConfig); err != nil {
			return err
		}
	}

	return nil
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

func createPatchConfigRequest(
	config *config.ConfigSettings,
	primaryResponse *model.ConfigResponse,
	replicaResponse *model.ConfigResponse,
	excludeDNSRecords []string,
) *model.PatchConfigRequest {
	patchConfig := model.PatchConfig{}

	if json := filterPatchConfigRequest(config.DNS, primaryResponse.Get("dns")); json != nil {
		if len(excludeDNSRecords) > 0 {
			var replicaDNS map[string]any
			if replicaResponse != nil {
				replicaDNS = replicaResponse.Get("dns")
			}
			json = mergeDNSRecords(json, replicaDNS, excludeDNSRecords)
		}
		patchConfig.DNS = json
	}
	if json := filterPatchConfigRequest(config.DHCP, primaryResponse.Get("dhcp")); json != nil {
		patchConfig.DHCP = json
	}
	if json := filterPatchConfigRequest(config.NTP, primaryResponse.Get("ntp")); json != nil {
		patchConfig.NTP = json
	}
	if json := filterPatchConfigRequest(config.Resolver, primaryResponse.Get("resolver")); json != nil {
		patchConfig.Resolver = json
	}
	if json := filterPatchConfigRequest(config.Database, primaryResponse.Get("database")); json != nil {
		patchConfig.Database = json
	}
	if json := filterPatchConfigRequest(config.Misc, primaryResponse.Get("misc")); json != nil {
		patchConfig.Misc = json
	}
	if json := filterPatchConfigRequest(config.Debug, primaryResponse.Get("debug")); json != nil {
		patchConfig.Debug = json
	}

	return &model.PatchConfigRequest{Config: patchConfig}
}

func mergeDNSRecords(primaryDNS, replicaDNS map[string]any, excludeList []string) map[string]any {
	if primaryDNS == nil {
		return nil
	}

	// Create a copy to avoid modifying the original primary map
	result := make(map[string]any)
	for k, v := range primaryDNS {
		result[k] = v
	}

	keysToMerge := []string{"hosts", "cnameRecords"}
	for _, key := range keysToMerge {
		primaryFiltered := filterDNSRecords(primaryDNS[key], excludeList, false)

		var replicaExcluded []any
		if replicaDNS != nil {
			replicaExcluded = filterDNSRecords(replicaDNS[key], excludeList, true)
		}

		result[key] = append(primaryFiltered, replicaExcluded...)
	}

	return result
}

func filterDNSRecords(listAny any, exclude []string, keepExcluded bool) []any {
	list, ok := listAny.([]any)
	if !ok {
		return nil
	}

	var filtered []any
	for _, item := range list {
		if shouldIncludeRecord(item, exclude, keepExcluded) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func shouldIncludeRecord(item any, exclude []string, keepExcluded bool) bool {
	itemStr := fmt.Sprintf("%v", item)
	isRecordExcluded := isExcluded(itemStr, exclude)

	if keepExcluded {
		return isRecordExcluded
	}
	return !isRecordExcluded
}

func isExcluded(itemStr string, exclude []string) bool {
	for _, ex := range exclude {
		if strings.Contains(itemStr, ex) {
			return true
		}
	}
	return false
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
