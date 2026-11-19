package e2e

import (
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/lovelaze/nebula-sync/internal/sync/leases"
)

const (
	restartTimeout = 60 * time.Second
	restartWindow  = 10 * time.Second
	restartPoll    = 100 * time.Millisecond
)

func (suite *testSuite) Test_SelectiveSync_MergeDHCPLeases() {
	nowSec := time.Now().Unix()
	laptop := fmt.Sprintf("%d aa:aa:aa:aa:aa:01 192.168.1.10 laptop *", nowSec+7200)
	staleLaptop := fmt.Sprintf("%d aa:aa:aa:aa:aa:01 192.168.1.10 laptop *", nowSec+3600)
	phone := fmt.Sprintf("%d aa:aa:aa:aa:aa:02 192.168.1.20 phone *", nowSec+7200)

	// The primary renewed the laptop; the replica handed out a lease to the
	// phone while it was the DHCP server and holds a stale copy of the laptop.
	suite.importLeases(suite.ph1, laptop+"\n")
	suite.importLeases(suite.ph2, staleLaptop+"\n"+phone+"\n")

	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(false))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(false))
	suite.T().Setenv("FULL_SYNC", "false")
	suite.T().Setenv("MERGE_DHCP_LEASES", "true")
	suite.T().Setenv("CLIENT_RETRY_DELAY_SECONDS", "3")

	s, err := service.Init()
	suite.Require().NoError(err)
	suite.Require().NoError(s.Run())

	expected := []model.DHCPLease{
		{Expires: nowSec + 7200, HWAddr: "aa:aa:aa:aa:aa:01", IP: "192.168.1.10", Name: "laptop", ClientID: "*"},
		{Expires: nowSec + 7200, HWAddr: "aa:aa:aa:aa:aa:02", IP: "192.168.1.20", Name: "phone", ClientID: "*"},
	}
	suite.Equal(expected, suite.waitForLeases(suite.ph1, 2), "primary")
	suite.Equal(expected, suite.waitForLeases(suite.ph2, 2), "replica")
}

// importLeases replaces the lease database of a Pi-hole and waits for FTL to restart.
func (suite *testSuite) importLeases(ph *PiHoleContainer, content string) {
	archive, err := leases.NewTeleporter([]byte(content))
	suite.Require().NoError(err)

	client := suite.client(ph)
	suite.Require().NoError(client.PostTeleporter(archive, &model.PostTeleporterRequest{DHCPLeases: true}))
	suite.waitForRestart(ph)
	suite.waitForLeases(ph, len(leases.Parse([]byte(content)).Leases))
}

// waitForRestart waits until FTL goes down after an import: it writes the
// imported files before answering and restarts afterwards.
func (suite *testSuite) waitForRestart(ph *PiHoleContainer) {
	deadline := time.Now().Add(restartWindow)
	for time.Now().Before(deadline) {
		if err := pihole.NewClient(ph.piHole(), http.DefaultClient).PostAuth(); err != nil {
			return
		}
		time.Sleep(restartPoll)
	}
}

// waitForLeases polls a Pi-hole, which may be restarting, until it reports count leases.
func (suite *testSuite) waitForLeases(ph *PiHoleContainer, count int) []model.DHCPLease {
	var found []model.DHCPLease

	suite.Require().EventuallyWithT(func(c *assert.CollectT) {
		client := pihole.NewClient(ph.piHole(), http.DefaultClient)
		if !assert.NoError(c, client.PostAuth()) {
			return
		}
		defer func() { _ = client.DeleteSession() }()

		var err error
		found, err = client.GetDHCPLeases()
		assert.NoError(c, err)
		assert.Len(c, found, count)
	}, restartTimeout, time.Second)

	return found
}

func (suite *testSuite) client(ph *PiHoleContainer) pihole.Client {
	client := pihole.NewClient(ph.piHole(), http.DefaultClient)
	suite.Require().NoError(client.PostAuth())
	suite.T().Cleanup(func() { _ = client.DeleteSession() })
	return client
}

func (c *PiHoleContainer) piHole() model.PiHole {
	return model.NewPiHole(c.ConnectionString(false), c.password)
}
