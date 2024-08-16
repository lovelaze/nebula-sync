package e2e

import (
	"context"
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type testSuite struct {
	suite.Suite
	piHole1 *PiHoleContainer
	piHole2 *PiHoleContainer
}

func (suite *testSuite) SetupTest() {
	ctx := context.Background()
	suite.piHole1 = RunPiHole(ctx, "foo1")
	suite.piHole2 = RunPiHole(ctx, "foo2")
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) Test_FullSync() {
	suite.T().Setenv("PRIMARY", suite.piHole1.EnvString())
	suite.T().Setenv("REPLICAS", suite.piHole2.EnvString())
	suite.T().Setenv("FULL_SYNC", "true")

	srv, err := service.Init()
	require.NoError(suite.T(), err)
	err = srv.Run()
	require.NoError(suite.T(), err)
}

func (suite *testSuite) Test_ManualSync() {
	suite.T().Setenv("PRIMARY", suite.piHole1.EnvString())
	suite.T().Setenv("REPLICAS", suite.piHole2.EnvString())
	suite.T().Setenv("FULL_SYNC", "false")

	srv, err := service.Init()
	require.NoError(suite.T(), err)
	err = srv.Run()
	require.NoError(suite.T(), err)
}
