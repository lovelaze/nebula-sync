package e2e

import (
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/stretchr/testify/require"
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
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString())
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString())
	suite.T().Setenv("FULL_SYNC", "true")

	s, err := service.Init()
	require.NoError(suite.T(), err)
	err = s.Run()
	require.NoError(suite.T(), err)
}

func (suite *testSuite) Test_ManualSync() {
	suite.T().Setenv("PRIMARY", suite.ph1.EnvString())
	suite.T().Setenv("REPLICAS", suite.ph2.EnvString())
	suite.T().Setenv("FULL_SYNC", "false")
	suite.T().Setenv("SYNC_CONFIG_DNS", "true")
	suite.T().Setenv("SYNC_CONFIG_DHCP", "true")
	suite.T().Setenv("SYNC_CONFIG_NTP", "true")
	suite.T().Setenv("SYNC_CONFIG_RESOLVER", "true")
	suite.T().Setenv("SYNC_CONFIG_DATABASE", "true")
	suite.T().Setenv("SYNC_CONFIG_MISC", "true")
	suite.T().Setenv("SYNC_CONFIG_DEBUG", "true")

	s, err := service.Init()
	require.NoError(suite.T(), err)
	err = s.Run()
	require.NoError(suite.T(), err)
}
