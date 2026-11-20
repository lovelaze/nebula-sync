package leases

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Unix(1_800_000_000, 0)

func at(offset time.Duration) int64 {
	return now.Add(offset).Unix()
}

func lease(expires int64, hwaddr, ip, name string) Lease {
	return Lease{Expires: expires, HWAddr: hwaddr, IP: ip, Name: name, ClientID: "*"}
}

const sampleFile = `1800003600 aa:aa:aa:aa:aa:01 192.168.1.10 laptop 01:aa:aa:aa:aa:aa:01
0 aa:aa:aa:aa:aa:02 192.168.1.11 * *
garbage line
duid 00:01:00:01:2c:2f:6e:2a:aa:aa:aa:aa:aa:ff
1800007200 1234 fd00::10 laptop 00:01:00:01:aa
1800007200 T99 fd00::11 * 00:01:00:01:bb
vendorclass 192.168.1.10 4d:53:46:54
`

func TestParseAndBytes_RoundTrip(t *testing.T) {
	file := Parse([]byte(sampleFile))

	assert.Equal(t, "00:01:00:01:2c:2f:6e:2a:aa:aa:aa:aa:aa:ff", file.DUID)
	require.Len(t, file.Leases, 4)
	assert.Equal(
		t,
		Lease{
			Expires:  1800003600,
			HWAddr:   "aa:aa:aa:aa:aa:01",
			IP:       "192.168.1.10",
			Name:     "laptop",
			ClientID: "01:aa:aa:aa:aa:aa:01",
		},
		file.Leases[0],
	)
	assert.True(t, file.Leases[2].IsV6())

	expected := `1800003600 aa:aa:aa:aa:aa:01 192.168.1.10 laptop 01:aa:aa:aa:aa:aa:01
0 aa:aa:aa:aa:aa:02 192.168.1.11 * *
duid 00:01:00:01:2c:2f:6e:2a:aa:aa:aa:aa:aa:ff
1800007200 1234 fd00::10 laptop 00:01:00:01:aa
1800007200 T99 fd00::11 * 00:01:00:01:bb
vendorclass 192.168.1.10 4d:53:46:54
`
	assert.Equal(t, expected, string(file.Bytes()))
}

func TestParse_Empty(t *testing.T) {
	file := Parse(nil)
	assert.Empty(t, file.Leases)
	assert.Empty(t, file.Bytes())
}

func TestLease_LaterThan(t *testing.T) {
	short, long, infinite := lease(at(time.Hour), "", "", ""), lease(at(2*time.Hour), "", "", ""), lease(0, "", "", "")

	assert.True(t, long.LaterThan(short))
	assert.False(t, short.LaterThan(long))
	assert.False(t, long.LaterThan(long))
	assert.True(t, infinite.LaterThan(long))
	assert.False(t, long.LaterThan(infinite))
	assert.False(t, infinite.LaterThan(infinite))
}

func TestMerge_ReplicaCopyIsNotAChange(t *testing.T) {
	primary := &File{Leases: []Lease{
		lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop"),
	}}
	// The primary renewed the lease after the last sync; the replica still has the old copy.
	copied := []Lease{lease(at(30*time.Minute), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")}
	replica := Source{Name: "replica", Leases: copied, Known: NewSnapshot(copied)}

	result := Merge(primary, []Source{replica}, now)

	assert.False(t, result.PrimaryChanged)
	assert.Empty(t, result.Adopted)
	assert.Equal(t, primary.Leases, result.File.Leases)
}

func TestMerge_AdoptsLeaseGrantedByReplica(t *testing.T) {
	primary := &File{DUID: "00:01", Leases: []Lease{
		lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop"),
	}}
	granted := lease(at(3*time.Hour), "aa:aa:aa:aa:aa:02", "192.168.1.20", "phone")
	replica := Source{Name: "replica", Leases: []Lease{primary.Leases[0], granted}, Known: NewSnapshot(primary.Leases)}

	result := Merge(primary, []Source{replica}, now)

	assert.True(t, result.PrimaryChanged)
	assert.Equal(t, []Adopted{{Source: "replica", Lease: granted}}, result.Adopted)
	assert.Empty(t, result.Dropped)
	assert.Equal(t, []Lease{primary.Leases[0], granted}, result.File.Leases)
	assert.Equal(t, "00:01", result.File.DUID)
}

func TestMerge_SameAddressLastExpiringWins(t *testing.T) {
	primaryLease := lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
	primary := &File{Leases: []Lease{primaryLease}}

	t.Run("replica renewal expires later", func(t *testing.T) {
		renewed := lease(at(5*time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{renewed}}}, now)

		assert.True(t, result.PrimaryChanged)
		assert.Equal(t, []Lease{renewed}, result.File.Leases)
		assert.Equal(t, []Lease{primaryLease}, result.Dropped)
	})

	t.Run("other client got the address later", func(t *testing.T) {
		other := lease(at(5*time.Hour), "bb:bb:bb:bb:bb:01", "192.168.1.10", "tv")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{other}}}, now)

		assert.Equal(t, []Lease{other}, result.File.Leases)
	})

	t.Run("replica lease expires earlier", func(t *testing.T) {
		older := lease(at(30*time.Minute), "bb:bb:bb:bb:bb:01", "192.168.1.10", "tv")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{older}}}, now)

		assert.False(t, result.PrimaryChanged)
		assert.Equal(t, []Lease{primaryLease}, result.File.Leases)
	})

	t.Run("tie keeps the primary lease", func(t *testing.T) {
		tie := primaryLease
		tie.Name = "*"
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{tie}}}, now)

		assert.False(t, result.PrimaryChanged)
		assert.Equal(t, []Lease{primaryLease}, result.File.Leases)
	})

	t.Run("infinite lease wins", func(t *testing.T) {
		infinite := lease(0, "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{infinite}}}, now)

		assert.Equal(t, []Lease{infinite}, result.File.Leases)
	})
}

func TestMerge_SameClientLastExpiringWins(t *testing.T) {
	old := lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
	primary := &File{Leases: []Lease{old}}

	t.Run("client moved to a new address on the replica", func(t *testing.T) {
		moved := lease(at(4*time.Hour), "AA:AA:AA:AA:AA:01", "192.168.1.50", "laptop")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{moved}}}, now)

		assert.True(t, result.PrimaryChanged)
		assert.Equal(t, []Lease{moved}, result.File.Leases)
		assert.Equal(t, []Lease{old}, result.Dropped)
	})

	t.Run("replica lease of the client is older", func(t *testing.T) {
		stale := lease(at(30*time.Minute), "aa:aa:aa:aa:aa:01", "192.168.1.50", "laptop")
		result := Merge(primary, []Source{{Name: "r", Leases: []Lease{stale}}}, now)

		assert.False(t, result.PrimaryChanged)
		assert.Equal(t, []Lease{old}, result.File.Leases)
	})

	t.Run("IPv6 leases with the same IAID do not conflict", func(t *testing.T) {
		v6Primary := &File{Leases: []Lease{lease(at(time.Hour), "1234", "fd00::10", "laptop")}}
		other := lease(at(2*time.Hour), "1234", "fd00::20", "phone")
		result := Merge(v6Primary, []Source{{Name: "r", Leases: []Lease{other}}}, now)

		assert.Len(t, result.File.Leases, 2)
	})
}

func TestMerge_DoesNotResurrectLeaseDroppedByPrimary(t *testing.T) {
	released := lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
	primary := &File{} // the client released its lease on the primary

	known := Source{Name: "r", Leases: []Lease{released}, Known: NewSnapshot([]Lease{released})}
	result := Merge(primary, []Source{known}, now)
	assert.False(t, result.PrimaryChanged)
	assert.Empty(t, result.File.Leases)

	// Without a snapshot (first run after start) every replica lease is a candidate.
	unknown := Source{Name: "r", Leases: []Lease{released}}
	result = Merge(primary, []Source{unknown}, now)
	assert.True(t, result.PrimaryChanged)
}

func TestMerge_DropsExpiredLeases(t *testing.T) {
	expired := lease(at(-time.Minute), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop")
	active := lease(at(time.Hour), "aa:aa:aa:aa:aa:02", "192.168.1.11", "phone")
	primary := &File{Leases: []Lease{expired, active}}

	result := Merge(primary, []Source{{Name: "r", Leases: []Lease{
		lease(at(-time.Hour), "aa:aa:aa:aa:aa:03", "192.168.1.12", "tv"),
	}}}, now)

	assert.False(t, result.PrimaryChanged, "an expired primary lease is not a change worth an import")
	assert.Equal(t, []Lease{active}, result.File.Leases)
}

func TestMerge_LeavesPrimaryDuplicatesAlone(t *testing.T) {
	// Two leases for one MAC on the primary must not trigger an import every run.
	primary := &File{Leases: []Lease{
		lease(at(time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.10", "laptop"),
		lease(at(2*time.Hour), "aa:aa:aa:aa:aa:01", "192.168.1.11", "laptop"),
	}}

	result := Merge(primary, nil, now)

	assert.False(t, result.PrimaryChanged)
	assert.Len(t, result.File.Leases, 2)
}

func TestMerge_ExtrasFollowPrimaryLeases(t *testing.T) {
	primary := Parse([]byte(`1800003600 aa:aa:aa:aa:aa:01 192.168.1.10 laptop *
1800003600 aa:aa:aa:aa:aa:02 192.168.1.11 phone *
vendorclass 192.168.1.10 4d:53
agent-info 192.168.1.11 01:02
`))
	superseding := lease(at(5*time.Hour), "bb:bb:bb:bb:bb:01", "192.168.1.11", "tv")

	result := Merge(primary, []Source{{Name: "r", Leases: []Lease{superseding}}}, now)

	assert.Equal(t, `1800003600 aa:aa:aa:aa:aa:01 192.168.1.10 laptop *
1800018000 bb:bb:bb:bb:bb:01 192.168.1.11 tv *
vendorclass 192.168.1.10 4d:53
`, string(result.File.Bytes()))
}

func TestMerge_OrdersByExpiry(t *testing.T) {
	infinite := lease(0, "aa:aa:aa:aa:aa:01", "192.168.1.1", "a")
	late := lease(at(2*time.Hour), "aa:aa:aa:aa:aa:02", "192.168.1.2", "b")
	early := lease(at(time.Hour), "aa:aa:aa:aa:aa:03", "192.168.1.3", "c")

	result := Merge(&File{Leases: []Lease{infinite, late, early}}, nil, now)

	assert.Equal(t, []Lease{early, late, infinite}, result.File.Leases)
}
