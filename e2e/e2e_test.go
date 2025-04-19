package e2e

import (
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/stretchr/testify/suite"
	"testing"
)

type testSuite struct {
	suite.Suite
	ph1 *PiHoleContainer
	ph2 *PiHoleContainer
}

func (suite *testSuite) SetupTest() {
	suite.ph1 = RunPiHole("foo1")
	suite.ph2 = RunPiHole("foo2")
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) Test_FullSync() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(false))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(false))
	suite.T().Setenv("FULL_SYNC", "true")
	suite.T().Setenv("RUN_GRAVITY", "true")

	s, err := service.Init()
	suite.Require().NoError(err)
	err = s.Run()
	suite.Require().NoError(err)
}

func (suite *testSuite) Test_FullSync_SSL() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(true))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(true))
	suite.T().Setenv("FULL_SYNC", "true")
	suite.T().Setenv("CLIENT_SKIP_TLS_VERIFICATION", "true")

	s, err := service.Init()
	suite.Require().NoError(err)
	err = s.Run()
	suite.Require().NoError(err)
}

func (suite *testSuite) Test_SelectiveSync() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(false))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(false))
	suite.T().Setenv("FULL_SYNC", "false")
	suite.T().Setenv("RUN_GRAVITY", "true")
	setAllManualConfig(suite)
	setAllManualGravity(suite)

	s, err := service.Init()
	suite.Require().NoError(err)
	err = s.Run()
	suite.Require().NoError(err)
}

func (suite *testSuite) Test_SelectiveSync_Include() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(false))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(false))
	suite.T().Setenv("FULL_SYNC", "false")
	setAllManualConfig(suite)
	setAllManualGravity(suite)

	suite.T().Setenv("SYNC_CONFIG_DNS_INCLUDE", "upstreams,blockESNI")
	suite.T().Setenv("SYNC_CONFIG_DHCP_INCLUDE", "active,start")
	suite.T().Setenv("SYNC_CONFIG_NTP_INCLUDE", "ipv4,sync")
	suite.T().Setenv("SYNC_CONFIG_RESOLVER_INCLUDE", "resolveIPv4,networkNames")
	suite.T().Setenv("SYNC_CONFIG_DATABASE_INCLUDE", "DBimport,maxDBdays")
	suite.T().Setenv("SYNC_CONFIG_MISC_INCLUDE", "nice,delay_startup")
	suite.T().Setenv("SYNC_CONFIG_DEBUG_INCLUDE", "database,networking")

	s, err := service.Init()
	suite.Require().NoError(err)
	err = s.Run()
	suite.Require().NoError(err)
}

func (suite *testSuite) Test_SelectiveSync_Exclude() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString(false))
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString(false))
	suite.T().Setenv("FULL_SYNC", "false")
	setAllManualConfig(suite)
	setAllManualGravity(suite)

	suite.T().Setenv("SYNC_CONFIG_DNS_EXCLUDE", "upstreams,blockESNI")
	suite.T().Setenv("SYNC_CONFIG_DHCP_EXCLUDE", "active,start")
	suite.T().Setenv("SYNC_CONFIG_NTP_EXCLUDE", "ipv4,sync")
	suite.T().Setenv("SYNC_CONFIG_RESOLVER_EXCLUDE", "resolveIPv4,networkNames")
	suite.T().Setenv("SYNC_CONFIG_DATABASE_EXCLUDE", "DBimport,maxDBdays")
	suite.T().Setenv("SYNC_CONFIG_MISC_EXCLUDE", "nice,delay_startup")
	suite.T().Setenv("SYNC_CONFIG_DEBUG_EXCLUDE", "database,networking")

	s, err := service.Init()
	suite.Require().NoError(err)
	err = s.Run()
	suite.Require().NoError(err)
}

func setAllManualConfig(suite *testSuite) {
	suite.T().Setenv("SYNC_CONFIG_DNS", "true")
	suite.T().Setenv("SYNC_CONFIG_DHCP", "true")
	suite.T().Setenv("SYNC_CONFIG_NTP", "true")
	suite.T().Setenv("SYNC_CONFIG_RESOLVER", "true")
	suite.T().Setenv("SYNC_CONFIG_DATABASE", "true")
	suite.T().Setenv("SYNC_CONFIG_MISC", "true")
	suite.T().Setenv("SYNC_CONFIG_DEBUG", "true")
}

func setAllManualGravity(suite *testSuite) {
	suite.T().Setenv("SYNC_GRAVITY_DHCP_LEASES", "true")
	suite.T().Setenv("SYNC_GRAVITY_GROUP", "true")
	suite.T().Setenv("SYNC_GRAVITY_AD_LIST", "true")
	suite.T().Setenv("SYNC_GRAVITY_AD_LIST_BY_GROUP", "true")
	suite.T().Setenv("SYNC_GRAVITY_DOMAIN_LIST", "true")
	suite.T().Setenv("SYNC_GRAVITY_DOMAIN_LIST_BY_GROUP", "true")
	suite.T().Setenv("SYNC_GRAVITY_CLIENT", "true")
	suite.T().Setenv("SYNC_GRAVITY_CLIENT_BY_GROUP", "true")
}
