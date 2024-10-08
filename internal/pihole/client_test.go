package pihole

import (
	"context"
	"fmt"
	"github.com/lovelaze/nebula-sync/e2e"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"testing"
)

const (
	apiPassword string = "test"
)

var (
	piHole = e2e.RunPiHole(apiPassword).Container
)

type clientTestSuite struct {
	suite.Suite
	client Client
}

func (suite *clientTestSuite) SetupTest() {
	client := createClient(piHole)
	err := client.Authenticate()
	require.NoError(suite.T(), err)
	suite.client = client
}

func TestClientIntegration(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (suite *clientTestSuite) TestClient_Authenticate() {
	err := suite.client.Authenticate()

	assert.NoError(suite.T(), err)
}

func (suite *clientTestSuite) TestClient_DeleteSession() {
	err := suite.client.DeleteSession()

	assert.NoError(suite.T(), err)
}

func (suite *clientTestSuite) TestClient_GetVersion() {
	version, err := suite.client.GetVersion()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), version)
}

func (suite *clientTestSuite) TestClient_GetTeleporter() {
	payload, err := suite.client.GetTeleporter()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), payload)
}

func (suite *clientTestSuite) TestClient_PostTeleporter() {
	payload, _ := suite.client.GetTeleporter()
	err := suite.client.PostTeleporter(payload, &model.PostTeleporterRequest{
		Config:     true,
		DHCPLeases: true,
		Gravity: model.PostGravityRequest{
			Group:             true,
			Adlist:            true,
			AdlistByGroup:     true,
			Domainlist:        true,
			DomainlistByGroup: true,
			Client:            true,
			ClientByGroup:     true,
		},
	})

	assert.NoError(suite.T(), err)
}

func (suite *clientTestSuite) TestClient_GetConfig() {
	conf, err := suite.client.GetConfig()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), conf)
}

func (suite *clientTestSuite) TestClient_PatchConfig() {
	request := model.PatchConfigRequest{
		Config: model.PatchConfig{
			DNS:      nil,
			DHCP:     nil,
			NTP:      nil,
			Resolver: nil,
			Database: nil,
			Misc:     nil,
			Debug:    nil,
		}}
	err := suite.client.PatchConfig(&request)

	assert.NoError(suite.T(), err)
}

func TestClient_String(t *testing.T) {
	piHole := model.NewPiHole("http://asdfasdf.com:1234", apiPassword)
	s := NewClient(piHole).String()

	assert.Equal(t, "http://asdfasdf.com:1234", s)
}

func TestClient_ApiPath(t *testing.T) {
	piHole := model.NewPiHole("http://asdfasdf.com:1234", apiPassword)
	c := NewClient(piHole)

	url := c.String()
	path := c.ApiPath("testing")
	expectedPath := fmt.Sprintf("%s/api/testing", url)

	assert.Equal(t, expectedPath, path)
}

func Test_auth_verify(t *testing.T) {
	a := auth{
		sid:      "",
		csrf:     "",
		validity: 0,
		valid:    false,
	}
	assert.Error(t, a.verify())

	a.valid = true
	assert.NoError(t, a.verify())
}

func createClient(container tc.Container) Client {
	apiPort, err := container.MappedPort(context.Background(), "80/tcp")
	if err != nil {
		panic(err)
	}

	host := fmt.Sprintf("http://localhost:%s", apiPort.Port())

	return NewClient(model.NewPiHole(host, apiPassword))
}
