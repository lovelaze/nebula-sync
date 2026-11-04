package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
)

func TestChangeTracker_due(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := newChangeTracker()
	tracker.now = func() time.Time { return now }

	due, reason := tracker.due("a", time.Hour)
	assert.True(t, due)
	assert.Equal(t, "first sync", reason)

	tracker.record("a")

	due, _ = tracker.due("a", time.Hour)
	assert.False(t, due)

	due, reason = tracker.due("b", time.Hour)
	assert.True(t, due)
	assert.Equal(t, "changes detected", reason)

	now = now.Add(time.Hour)
	due, reason = tracker.due("a", time.Hour)
	assert.True(t, due)
	assert.Equal(t, "max age reached", reason)

	due, _ = tracker.due("a", 0)
	assert.False(t, due, "max age 0 never forces a sync")
}

func TestTarget_fingerprint_ignoresVolatileFields(t *testing.T) {
	primary := piholemock.NewClient(t)
	tgt := &target{Primary: primary, changes: newChangeTracker()}
	settings := &config.GravitySettings{Adlist: true}

	list := func(address string, number, status int) map[string]any {
		return map[string]any{"took": 0.001, "lists": []any{
			map[string]any{
				"id":      1.0,
				"address": address,
				"type":    "block",
				"enabled": true,
				"comment": nil,
				"groups":  []any{0.0},
				"date_updated": 1700000000.0 + float64(
					number,
				),
				"number":          float64(number),
				"invalid_domains": 0.0,
				"status":          float64(status),
			},
		}}
	}

	primary.EXPECT().GetJSON("lists").Once().Return(list("https://a.example/list", 100, 1), nil)
	primary.EXPECT().GetJSON("lists").Once().Return(list("https://a.example/list", 250, 2), nil)
	primary.EXPECT().GetJSON("lists").Once().Return(list("https://b.example/list", 100, 1), nil)

	first, err := tgt.fingerprint(settings, nil)
	require.NoError(t, err)
	afterGravity, err := tgt.fingerprint(settings, nil)
	require.NoError(t, err)
	changed, err := tgt.fingerprint(settings, nil)
	require.NoError(t, err)

	assert.Equal(t, first, afterGravity, "gravity counters and status are not a change")
	assert.NotEqual(t, first, changed, "a different adlist address is a change")
}

func TestEnabledGravityResources(t *testing.T) {
	assert.Empty(t, enabledGravityResources(nil))
	assert.Empty(t, enabledGravityResources(&config.GravitySettings{DHCPLeases: true}))

	resources := enabledGravityResources(newFullSyncGravitySettings())
	endpoints := make([]string, 0, len(resources))
	for _, r := range resources {
		endpoints = append(endpoints, r.endpoint)
	}
	assert.Equal(t, []string{"groups", "lists", "domains", "clients"}, endpoints)
}

func onChangeSettings() *config.Sync {
	return &config.Sync{
		SyncOnChange:       true,
		SyncOnChangeMaxAge: 24 * time.Hour,
		GravitySettings:    &config.GravitySettings{Group: true},
		ConfigSettings: &config.ConfigSettings{
			DNS:       config.NewConfigSetting(true, nil, nil),
			DHCP:      config.NewConfigSetting(false, nil, nil),
			NTP:       config.NewConfigSetting(false, nil, nil),
			Resolver:  config.NewConfigSetting(false, nil, nil),
			Database:  config.NewConfigSetting(false, nil, nil),
			Webserver: config.NewConfigSetting(false, nil, nil),
			Files:     config.NewConfigSetting(false, nil, nil),
			Misc:      config.NewConfigSetting(false, nil, nil),
			Debug:     config.NewConfigSetting(false, nil, nil),
		},
	}
}

func groupsResponse(name string) map[string]any {
	return map[string]any{"took": 0.002, "groups": []any{
		map[string]any{"id": 0.0, "name": name, "enabled": true, "comment": nil, "date_modified": 1700000000.0},
	}}
}

func expectSessions(primary, replica *piholemock.Client, times int) {
	primary.EXPECT().PostAuth().Times(times).Return(nil)
	replica.EXPECT().PostAuth().Times(times).Return(nil)
	primary.EXPECT().DeleteSession().Times(times).Return(nil)
	replica.EXPECT().DeleteSession().Times(times).Return(nil)
}

func expectSync(primary, replica *piholemock.Client) {
	primary.EXPECT().GetTeleporter().Once().Return([]byte{}, nil)
	replica.EXPECT().PostTeleporter(mock.Anything, mock.Anything).Once().Return(nil)
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	replica.EXPECT().PatchConfig(mock.Anything).Once().Return(nil)
}

func TestTarget_SelectiveSync_onChange(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := NewTarget(primary, []pihole.Client{replica})
	settings := onChangeSettings()

	expectSessions(primary, replica, 3)

	// 1st run: no previous state, so it syncs
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	primary.EXPECT().GetJSON("groups").Once().Return(groupsResponse("Default"), nil)
	expectSync(primary, replica)
	require.NoError(t, target.SelectiveSync(settings))

	// 2nd run: same state (only "took" differs), so it skips
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	unchanged := groupsResponse("Default")
	unchanged["took"] = 0.009
	primary.EXPECT().GetJSON("groups").Once().Return(unchanged, nil)
	require.ErrorIs(t, target.SelectiveSync(settings), ErrNoChanges)

	// 3rd run: a group was renamed, so it syncs
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	primary.EXPECT().GetJSON("groups").Once().Return(groupsResponse("IoT"), nil)
	expectSync(primary, replica)
	require.NoError(t, target.SelectiveSync(settings))
}

func TestTarget_SelectiveSync_onChange_failedSyncIsRetried(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := NewTarget(primary, []pihole.Client{replica})
	settings := onChangeSettings()

	expectSessions(primary, replica, 2)

	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	primary.EXPECT().GetJSON("groups").Once().Return(groupsResponse("Default"), nil)
	primary.EXPECT().GetTeleporter().Once().Return(nil, errors.New("boom"))
	require.Error(t, target.SelectiveSync(settings))

	// same state, but the last sync failed, so it is not recorded and syncs again
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	primary.EXPECT().GetJSON("groups").Once().Return(groupsResponse("Default"), nil)
	expectSync(primary, replica)
	require.NoError(t, target.SelectiveSync(settings))
}

func TestTarget_SelectiveSync_onChange_fingerprintErrorSyncs(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := NewTarget(primary, []pihole.Client{replica})

	expectSessions(primary, replica, 1)
	primary.EXPECT().GetConfig().Once().Return(emptyConfigResponse(), nil)
	primary.EXPECT().GetJSON("groups").Once().Return(nil, errors.New("unreachable"))
	expectSync(primary, replica)

	require.NoError(t, target.SelectiveSync(onChangeSettings()))
}
