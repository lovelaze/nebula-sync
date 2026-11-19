package sync

import (
	"archive/zip"
	"bytes"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	piholemock "github.com/lovelaze/nebula-sync/internal/mocks/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/sync/leases"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
)

func teleporterWithLeases(t *testing.T, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, data := range map[string]string{"etc/pihole/pihole.toml": "toml", leases.TeleporterPath: content} {
		entry, err := writer.Create(name)
		require.NoError(t, err)
		_, err = entry.Write([]byte(data))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	return buf.Bytes()
}

func leasesIn(t *testing.T, archive []byte) string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	for _, file := range reader.File {
		if file.Name == leases.TeleporterPath {
			content, err := file.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(content)
			require.NoError(t, err)
			return string(data)
		}
	}

	t.Fatal("no leases in archive")
	return ""
}

func TestTarget_mergeDHCPLeases_failoverThenSteadyState(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := &target{Primary: primary, Replicas: []pihole.Client{replica}}

	gravitySettings := &config.GravitySettings{Group: true}
	baseSec := time.Now().Unix()
	expires := strconv.FormatInt(baseSec+3600, 10)
	laterExpires := baseSec + 7200

	primaryLeases := expires + " aa:aa:aa:aa:aa:01 192.168.1.10 laptop *\nduid 00:01\n"
	primaryTeleporter := teleporterWithLeases(t, primaryLeases)
	merged := expires + " aa:aa:aa:aa:aa:01 192.168.1.10 laptop *\n" +
		strconv.FormatInt(laterExpires, 10) + " aa:aa:aa:aa:aa:02 192.168.1.20 phone *\n" +
		"duid 00:01\n"

	replica.EXPECT().String().Return("http://replica")
	primary.EXPECT().String().Return("http://primary").Maybe()

	// Run 1: the replica handed out a lease while the primary was down.
	primary.EXPECT().GetTeleporter().Twice().Return(primaryTeleporter, nil)
	replica.EXPECT().GetDHCPLeases().Once().Return([]model.DHCPLease{
		{Expires: laterExpires, HWAddr: "aa:aa:aa:aa:aa:02", IP: "192.168.1.20", Name: "phone", ClientID: "*"},
	}, nil)
	replica.EXPECT().PostTeleporter(mock.Anything, mock.Anything).
		Run(func(payload []byte, request *model.PostTeleporterRequest) {
			assert.Equal(t, merged, leasesIn(t, payload))
			assert.True(t, request.DHCPLeases, "merged leases must be imported")
			assert.True(t, request.Gravity.Group, "gravity settings are kept")
		}).Return(nil).Once()

	merge, err := target.syncTeleporters(gravitySettings, true)
	require.NoError(t, err)
	require.True(t, merge.result.PrimaryChanged)

	primary.EXPECT().PostTeleporter(mock.Anything, mock.Anything).
		Run(func(payload []byte, request *model.PostTeleporterRequest) {
			assert.Equal(t, merged, leasesIn(t, payload))
			assert.Equal(t, &model.PostTeleporterRequest{DHCPLeases: true}, request)
		}).Return(nil).Once()

	require.NoError(t, target.updatePrimaryLeases(merge))

	// Run 2: the primary now has every lease and the replica holds the copy
	// written in run 1, so the primary is left alone.
	mergedTeleporter := teleporterWithLeases(t, merged)
	primary.EXPECT().GetTeleporter().Once().Return(mergedTeleporter, nil)
	replica.EXPECT().GetDHCPLeases().Once().Return([]model.DHCPLease{
		{Expires: baseSec + 3600, HWAddr: "aa:aa:aa:aa:aa:01", IP: "192.168.1.10", Name: "laptop", ClientID: "*"},
		{Expires: laterExpires, HWAddr: "aa:aa:aa:aa:aa:02", IP: "192.168.1.20", Name: "phone", ClientID: "*"},
	}, nil)
	replica.EXPECT().PostTeleporter(mock.Anything, mock.Anything).Once().Return(nil)

	merge, err = target.syncTeleporters(gravitySettings, true)
	require.NoError(t, err)
	assert.False(t, merge.result.PrimaryChanged)
	require.NoError(t, target.updatePrimaryLeases(merge))
}

func TestTarget_updatePrimaryLeases_remergesWithCurrentPrimary(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := &target{Primary: primary, Replicas: []pihole.Client{replica}}

	expiresSec := time.Now().Add(time.Hour).Unix()
	replicaLease := model.DHCPLease{
		Expires:  expiresSec,
		HWAddr:   "aa:aa:aa:aa:aa:02",
		IP:       "192.168.1.20",
		Name:     "phone",
		ClientID: "*",
	}
	replica.EXPECT().String().Return("http://replica")
	replica.EXPECT().GetDHCPLeases().Once().Return([]model.DHCPLease{replicaLease}, nil)

	merge, err := target.mergeDHCPLeases(teleporterWithLeases(t, ""))
	require.NoError(t, err)
	require.True(t, merge.result.PrimaryChanged)

	// Meanwhile the primary renewed the lease itself, later than the replica.
	renewed := strconv.FormatInt(expiresSec+600, 10) + " aa:aa:aa:aa:aa:02 192.168.1.20 phone *\n"
	primary.EXPECT().GetTeleporter().Once().Return(teleporterWithLeases(t, renewed), nil)

	require.NoError(t, target.updatePrimaryLeases(merge))
	primary.AssertNotCalled(t, "PostTeleporter", mock.Anything, mock.Anything)
}

func TestTarget_updatePrimaryLeases_noMerge(t *testing.T) {
	target := &target{Primary: piholemock.NewClient(t)}

	require.NoError(t, target.updatePrimaryLeases(nil))
}

func TestTarget_mergeDHCPLeases_retriesAfterFailedPrimaryImport(t *testing.T) {
	primary := piholemock.NewClient(t)
	replica := piholemock.NewClient(t)
	target := &target{Primary: primary, Replicas: []pihole.Client{replica}}

	expiresSec := time.Now().Add(time.Hour).Unix()
	granted := model.DHCPLease{
		Expires:  expiresSec,
		HWAddr:   "aa:aa:aa:aa:aa:02",
		IP:       "192.168.1.20",
		Name:     "phone",
		ClientID: "*",
	}
	primaryTeleporter := teleporterWithLeases(t, "")

	replica.EXPECT().String().Return("http://replica")
	primary.EXPECT().String().Return("http://primary").Maybe()
	primary.EXPECT().GetTeleporter().Return(primaryTeleporter, nil)
	replica.EXPECT().GetDHCPLeases().Return([]model.DHCPLease{granted}, nil)
	replica.EXPECT().PostTeleporter(mock.Anything, mock.Anything).Return(nil)

	merge, err := target.syncTeleporters(nil, true)
	require.NoError(t, err)
	require.Len(t, merge.adopted, 1)

	primary.EXPECT().PostTeleporter(mock.Anything, mock.Anything).Return(assert.AnError).Times(retry.AttemptsPostTeleporter)
	require.Error(t, target.updatePrimaryLeases(merge))

	// The replica still holds the lease nebula-sync wrote back to it, but the
	// primary never got it: the next sync must offer it again.
	merge, err = target.syncTeleporters(nil, true)
	require.NoError(t, err)
	assert.True(t, merge.result.PrimaryChanged)
	assert.Len(t, merge.adopted, 1)
}
