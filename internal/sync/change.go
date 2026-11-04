package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lovelaze/nebula-sync/internal/config"
)

// ErrNoChanges is returned by a sync when SYNC_ON_CHANGE is enabled and nothing
// that would be synced has changed on the primary since the last successful sync.
var ErrNoChanges = errors.New("no changes on primary since last sync")

// gravityResource is a Pi-hole API list endpoint and the fields of each item that
// describe its configuration. Fields that Pi-hole updates on its own (gravity
// counters, status, resolved client names, timings) are left out, so they do not
// count as a change.
type gravityResource struct {
	endpoint string
	fields   []string
}

const (
	fieldID      = "id"
	fieldEnabled = "enabled"
	fieldComment = "comment"
	fieldGroups  = endpointGroups
	fieldType    = "type"

	endpointGroups  = "groups"
	endpointLists   = "lists"
	endpointDomains = "domains"
	endpointClients = "clients"
)

var (
	groupsResource = gravityResource{
		endpoint: endpointGroups,
		fields:   []string{fieldID, "name", fieldEnabled, fieldComment},
	}
	listsResource = gravityResource{
		endpoint: endpointLists,
		fields:   []string{fieldID, "address", fieldType, fieldEnabled, fieldComment, fieldGroups},
	}
	domainsResource = gravityResource{
		endpoint: endpointDomains,
		fields:   []string{fieldID, "domain", fieldType, "kind", fieldEnabled, fieldComment, fieldGroups},
	}
	clientsResource = gravityResource{
		endpoint: endpointClients,
		fields:   []string{fieldID, "client", fieldComment, fieldGroups},
	}
)

// changeTracker remembers the fingerprint of the primary at the last successful
// sync. It is held in memory only, so the first sync after a start always runs.
type changeTracker struct {
	fingerprint string
	lastSync    time.Time
	now         func() time.Time
}

func newChangeTracker() *changeTracker {
	return &changeTracker{now: time.Now}
}

// due reports whether a sync should run for the given fingerprint.
func (c *changeTracker) due(fingerprint string, maxAge time.Duration) (bool, string) {
	switch {
	case c.fingerprint == "":
		return true, "first sync"
	case fingerprint != c.fingerprint:
		return true, "changes detected"
	case maxAge > 0 && c.now().Sub(c.lastSync) >= maxAge:
		return true, "max age reached"
	default:
		return false, ""
	}
}

func (c *changeTracker) record(fingerprint string) {
	c.fingerprint = fingerprint
	c.lastSync = c.now()
}

// fingerprint hashes everything a sync with these settings would copy from the
// primary: the filtered config patch and the enabled gravity tables. DHCP leases
// are not included, as they change constantly.
func (target *target) fingerprint(
	gravitySettings *config.GravitySettings,
	configSettings *config.ConfigSettings,
) (string, error) {
	state := map[string]any{}

	if configSettings != nil {
		configResponse, err := target.Primary.GetConfig()
		if err != nil {
			return "", err
		}
		state["config"] = createPatchConfigRequest(configSettings, configResponse)
	}

	for _, resource := range enabledGravityResources(gravitySettings) {
		items, err := target.gravityItems(resource)
		if err != nil {
			return "", err
		}
		state[resource.endpoint] = items
	}

	// encoding/json sorts map keys, so equal state always gives equal bytes
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func enabledGravityResources(settings *config.GravitySettings) []gravityResource {
	if settings == nil {
		return nil
	}

	var resources []gravityResource
	if settings.Group {
		resources = append(resources, groupsResource)
	}
	if settings.Adlist || settings.AdlistByGroup {
		resources = append(resources, listsResource)
	}
	if settings.Domainlist || settings.DomainlistByGroup {
		resources = append(resources, domainsResource)
	}
	if settings.Client || settings.ClientByGroup {
		resources = append(resources, clientsResource)
	}
	return resources
}

func (target *target) gravityItems(resource gravityResource) ([]map[string]any, error) {
	response, err := target.Primary.GetJSON(resource.endpoint)
	if err != nil {
		return nil, err
	}

	raw, ok := response[resource.endpoint].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response from %s: missing %q array", resource.endpoint, resource.endpoint)
	}

	items := make([]map[string]any, 0, len(raw))
	for _, entry := range raw {
		item, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unexpected item in %s response", resource.endpoint)
		}
		kept := make(map[string]any, len(resource.fields))
		for _, field := range resource.fields {
			kept[field] = item[field]
		}
		items = append(items, kept)
	}
	return items, nil
}
