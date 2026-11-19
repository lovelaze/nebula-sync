package sync

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/sync/leases"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
)

// leaseMerge is the DHCP lease merge of one sync run.
type leaseMerge struct {
	// sources are the leases read from the replicas.
	sources []leases.Source
	result  *leases.Result
	// adopted are the replica leases the primary lacked when the merge was
	// written to the replicas.
	adopted []leases.Lease
}

// run merges the replica leases with the leases in the primary's teleporter archive.
func (merge *leaseMerge) run(primaryTeleporter []byte) error {
	raw, err := leases.FromTeleporter(primaryTeleporter)
	if err != nil {
		return err
	}

	merge.result = leases.Merge(leases.Parse(raw), merge.sources, time.Now())
	return nil
}

// mergeDHCPLeases reads the leases of every replica and merges them with the
// primary's, so leases a replica handed out while it was the DHCP server are
// kept once the primary is back.
func (target *target) mergeDHCPLeases(primaryTeleporter []byte) (*leaseMerge, error) {
	log.Info().Msg("Merging DHCP leases...")

	merge := &leaseMerge{sources: make([]leases.Source, 0, len(target.Replicas))}
	for _, replica := range target.Replicas {
		var replicaLeases []model.DHCPLease
		if err := retry.Fixed(func() error {
			var err error
			replicaLeases, err = replica.GetDHCPLeases()
			return err
		}, retry.AttemptsGetDHCPLeases); err != nil {
			return nil, err
		}

		merge.sources = append(merge.sources, leases.Source{
			Name:   replica.String(),
			Leases: fromModel(replicaLeases),
			Known:  target.leaseSnapshots[replica.String()],
		})
	}

	if err := merge.run(primaryTeleporter); err != nil {
		return nil, err
	}

	for _, adopted := range merge.result.Adopted {
		merge.adopted = append(merge.adopted, adopted.Lease)
		log.Info().
			Str("replica", adopted.Source).
			Str("ip", adopted.Lease.IP).
			Str("hwaddr", adopted.Lease.HWAddr).
			Str("name", adopted.Lease.Name).
			Int64("expires", adopted.Lease.Expires).
			Msg("Adopting DHCP lease handed out by replica")
	}
	for _, dropped := range merge.result.Dropped {
		log.Debug().
			Str("ip", dropped.IP).
			Str("hwaddr", dropped.HWAddr).
			Str("name", dropped.Name).
			Int64("expires", dropped.Expires).
			Msg("Primary DHCP lease superseded by a replica lease")
	}

	return merge, nil
}

// rememberLeases records the merged leases written to a replica, so the next
// merge does not mistake them for leases the replica handed out. Adopted
// leases are left out until the primary has them (see confirmAdopted): if the
// primary import fails, the next sync offers them again.
func (target *target) rememberLeases(replica pihole.Client, merge *leaseMerge) {
	if target.leaseSnapshots == nil {
		target.leaseSnapshots = map[string]leases.Snapshot{}
	}

	snapshot := leases.NewSnapshot(merge.result.File.Leases)
	for key := range leases.NewSnapshot(merge.adopted) {
		delete(snapshot, key)
	}
	target.leaseSnapshots[replica.String()] = snapshot
}

// confirmAdopted records the adopted leases as written once the primary has them.
func (target *target) confirmAdopted(merge *leaseMerge) {
	for _, snapshot := range target.leaseSnapshots {
		snapshot.Add(merge.adopted)
	}
}

// updatePrimaryLeases imports the merged leases on the primary if it lacks
// some of them. The import restarts the primary's FTL, so it only happens
// when a replica handed out leases, and it must be the last step of a sync.
func (target *target) updatePrimaryLeases(merge *leaseMerge) error {
	if merge == nil || len(merge.adopted) == 0 {
		return nil
	}

	// The primary may have granted leases since the merge: merge again with
	// its current leases so the import does not drop them.
	conf, err := target.Primary.GetTeleporter()
	if err != nil {
		return err
	}
	if err := merge.run(conf); err != nil {
		return err
	}
	if !merge.result.PrimaryChanged {
		target.confirmAdopted(merge)
		return nil
	}

	archive, err := leases.NewTeleporter(merge.result.File.Bytes())
	if err != nil {
		return err
	}

	log.Info().
		Int("adopted", len(merge.result.Adopted)).
		Int("superseded", len(merge.result.Dropped)).
		Msgf("Importing merged DHCP leases on primary %s (Pi-hole FTL restarts)", target.Primary.String())

	request := &model.PostTeleporterRequest{DHCPLeases: true}
	if err := retry.Fixed(func() error {
		return target.Primary.PostTeleporter(archive, request)
	}, retry.AttemptsPostTeleporter); err != nil {
		return err
	}

	target.confirmAdopted(merge)
	return nil
}

func fromModel(modelLeases []model.DHCPLease) []leases.Lease {
	result := make([]leases.Lease, 0, len(modelLeases))
	for _, lease := range modelLeases {
		result = append(result, leases.Lease{
			Expires:  lease.Expires,
			HWAddr:   lease.HWAddr,
			IP:       lease.IP,
			Name:     lease.Name,
			ClientID: lease.ClientID,
		})
	}
	return result
}
