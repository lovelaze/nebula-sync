// Package leases reads, merges and writes the dnsmasq lease database Pi-hole
// keeps in /etc/pihole/dhcp.leases, so that leases handed out by a replica
// while it acted as DHCP server are not lost when the primary takes over again.
package leases

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// leaseFields is the number of fields of a lease record.
const leaseFields = 5

// Lease is one record of the lease database:
// "<expiry> <hwaddr|iaid> <address> <hostname|*> <clientid|*>".
type Lease struct {
	// Expires is a unix timestamp, 0 means the lease never expires.
	Expires int64
	// HWAddr is the MAC address for IPv4 leases and the IAID for IPv6 leases.
	HWAddr   string
	IP       string
	Name     string
	ClientID string
}

func (l Lease) String() string {
	return fmt.Sprintf("%d %s %s %s %s", l.Expires, l.HWAddr, l.IP, orStar(l.Name), orStar(l.ClientID))
}

// IsV6 reports whether the lease is for an IPv6 address.
func (l Lease) IsV6() bool {
	return strings.Contains(l.IP, ":")
}

// Expired reports whether the lease has run out at the given time.
func (l Lease) Expired(now time.Time) bool {
	return l.Expires != 0 && l.Expires <= now.Unix()
}

// LaterThan reports whether l expires strictly after other.
func (l Lease) LaterThan(other Lease) bool {
	switch {
	case l.Expires == 0:
		return other.Expires != 0
	case other.Expires == 0:
		return false
	default:
		return l.Expires > other.Expires
	}
}

// sameClient reports whether two IPv4 leases belong to the same MAC address.
func (l Lease) sameClient(other Lease) bool {
	return !l.IsV6() && !other.IsV6() && hasMAC(l) && strings.EqualFold(l.HWAddr, other.HWAddr)
}

// hasMAC is false for "*" or an empty hardware address, which do not identify a client.
func hasMAC(l Lease) bool {
	return strings.Contains(l.HWAddr, ":")
}

// key identifies a lease as granted: a renewal changes the expiry, a new
// grant changes the address or client.
type key struct {
	ip      string
	hwaddr  string
	expires int64
}

func (l Lease) key() key {
	return key{ip: l.IP, hwaddr: strings.ToLower(l.HWAddr), expires: l.Expires}
}

// extra is a "vendorclass" or "agent-info" record, attached to a lease by address.
type extra struct {
	kind string
	ip   string
	data string
}

// File is a parsed lease database.
type File struct {
	// DUID is the server DUID dnsmasq uses for DHCPv6, empty if none.
	DUID   string
	Leases []Lease
	extras []extra
}

// Parse reads a lease database. Lines that are not a lease, the server DUID
// or a lease extra are skipped, the same way dnsmasq skips them.
func Parse(data []byte) *File {
	file := &File{}

	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)

		switch {
		case len(fields) == 2 && fields[0] == "duid":
			file.DUID = fields[1]
		case len(fields) == 3 && (fields[0] == "vendorclass" || fields[0] == "agent-info"):
			file.extras = append(file.extras, extra{kind: fields[0], ip: fields[1], data: fields[2]})
		case len(fields) == leaseFields:
			expires, err := strconv.ParseInt(fields[0], 10, 64)
			if err != nil || expires < 0 {
				continue
			}
			file.Leases = append(file.Leases, Lease{
				Expires:  expires,
				HWAddr:   fields[1],
				IP:       fields[2],
				Name:     fields[3],
				ClientID: fields[4],
			})
		}
	}

	return file
}

// Bytes writes the lease database in the layout dnsmasq uses: IPv4 leases,
// the server DUID, IPv6 leases, then lease extras.
func (f *File) Bytes() []byte {
	lines := make([]string, 0, len(f.Leases)+len(f.extras)+1)
	ips := make(map[string]bool, len(f.Leases))

	for _, lease := range f.Leases {
		ips[lease.IP] = true
		if !lease.IsV6() {
			lines = append(lines, lease.String())
		}
	}

	if f.DUID != "" {
		lines = append(lines, "duid "+f.DUID)
	}
	for _, lease := range f.Leases {
		if lease.IsV6() {
			lines = append(lines, lease.String())
		}
	}

	for _, e := range f.extras {
		if ips[e.ip] {
			lines = append(lines, e.kind+" "+e.ip+" "+e.data)
		}
	}

	if len(lines) == 0 {
		return nil
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// Snapshot records the leases last written to a replica. The next merge uses
// it to tell the leases a replica handed out or renewed itself from the copies
// it received; the copies must not resurrect leases the primary has dropped.
type Snapshot map[key]struct{}

// NewSnapshot records the given leases.
func NewSnapshot(leases []Lease) Snapshot {
	snapshot := make(Snapshot, len(leases))
	for _, lease := range leases {
		snapshot[lease.key()] = struct{}{}
	}
	return snapshot
}

// Add records more leases.
func (s Snapshot) Add(leases []Lease) {
	for _, lease := range leases {
		s[lease.key()] = struct{}{}
	}
}

// Source is the lease set read from one replica.
type Source struct {
	Name   string
	Leases []Lease
	// Known holds the leases last written to this replica. Nil means nothing
	// is known (e.g. right after nebula-sync started): every lease of the
	// replica is then a candidate.
	Known Snapshot
}

// Adopted is a lease taken from a replica.
type Adopted struct {
	Source string
	Lease  Lease
}

// Result is the outcome of a merge.
type Result struct {
	// File is the merged database: the primary's DUID and extras, every active lease.
	File *File
	// PrimaryChanged is true when File differs from the primary's active
	// leases, i.e. the primary must import it.
	PrimaryChanged bool
	// Adopted lists the leases taken from replicas.
	Adopted []Adopted
	// Dropped lists the primary leases replaced by a replica lease.
	Dropped []Lease
}

// Merge combines the primary's leases with the leases replicas handed out
// themselves. When two leases conflict, the one that expires last wins:
//   - leases for the same address conflict;
//   - IPv4 leases for the same MAC address on different addresses conflict,
//     since a client holds one lease at a time.
//
// On a tie the primary's lease is kept. Expired leases are dropped. A replica
// lease that was last written to it by nebula-sync (see Snapshot) is a copy,
// not a lease the replica granted, and is ignored.
func Merge(primary *File, replicas []Source, now time.Time) *Result {
	m := newMerger(primary, now)
	for _, source := range replicas {
		for _, candidate := range source.Leases {
			if _, known := source.Known[candidate.key()]; !known {
				m.offer(source.Name, candidate)
			}
		}
	}

	return m.result(primary)
}

// merger holds the merged leases by address while replica leases are offered.
type merger struct {
	now           time.Time
	byIP          map[string]Lease
	primaryActive map[string]Lease
	adopted       map[string]Adopted
}

func newMerger(primary *File, now time.Time) *merger {
	m := &merger{
		now:           now,
		byIP:          map[string]Lease{},
		primaryActive: map[string]Lease{},
		adopted:       map[string]Adopted{},
	}

	for _, lease := range primary.Leases {
		if _, dup := m.byIP[lease.IP]; dup || lease.Expired(now) {
			continue
		}
		m.byIP[lease.IP] = lease
		m.primaryActive[lease.IP] = lease
	}

	return m
}

// offer adds a replica lease if it wins over the leases it conflicts with.
func (m *merger) offer(source string, candidate Lease) {
	if candidate.Expired(m.now) {
		return
	}
	if current, ok := m.byIP[candidate.IP]; ok && !candidate.LaterThan(current) {
		return
	}

	if hasMAC(candidate) && !candidate.IsV6() {
		others := m.sameClient(candidate)
		for _, other := range others {
			if !candidate.LaterThan(other) {
				return // the client holds a newer lease on another address
			}
		}
		for _, other := range others {
			delete(m.byIP, other.IP)
			delete(m.adopted, other.IP)
		}
	}

	m.byIP[candidate.IP] = candidate
	m.adopted[candidate.IP] = Adopted{Source: source, Lease: candidate}
}

// sameClient returns the leases of the candidate's client on other addresses.
func (m *merger) sameClient(candidate Lease) []Lease {
	var others []Lease
	for ip, lease := range m.byIP {
		if ip != candidate.IP && lease.sameClient(candidate) {
			others = append(others, lease)
		}
	}
	return others
}

func (m *merger) result(primary *File) *Result {
	result := &Result{File: &File{DUID: primary.DUID}}

	for _, a := range m.adopted {
		result.Adopted = append(result.Adopted, a)
	}
	slices.SortFunc(result.Adopted, func(a, b Adopted) int { return cmp.Compare(a.Lease.IP, b.Lease.IP) })

	for ip, lease := range m.primaryActive {
		if merged, ok := m.byIP[ip]; !ok || merged != lease {
			result.Dropped = append(result.Dropped, lease)
		}
	}
	sortLeases(result.Dropped)
	result.PrimaryChanged = len(result.Adopted) > 0 || len(result.Dropped) > 0

	for _, lease := range m.byIP {
		result.File.Leases = append(result.File.Leases, lease)
	}
	sortLeases(result.File.Leases)

	// Keep the primary's extras only for the leases that are still the primary's.
	for _, e := range primary.extras {
		if lease, ok := m.byIP[e.ip]; ok && lease == m.primaryActive[e.ip] {
			result.File.extras = append(result.File.extras, e)
		}
	}

	return result
}

// sortLeases orders leases by expiry, never-expiring last, then by address.
// dnsmasq gives a hostname to the last lease it loads with that name, so the
// lease that expires last keeps it.
func sortLeases(leases []Lease) {
	slices.SortFunc(leases, func(a, b Lease) int {
		switch {
		case b.LaterThan(a):
			return -1
		case a.LaterThan(b):
			return 1
		default:
			return cmp.Compare(a.IP, b.IP)
		}
	})
}

func orStar(s string) string {
	if s == "" {
		return "*"
	}
	return s
}
