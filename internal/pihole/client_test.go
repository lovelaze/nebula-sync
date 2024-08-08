package pihole

import (
	"context"
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"testing"
	"time"
)

const (
	dockerImage string = "pihole/pihole:development-v6"
	apiPassword string = "test"
)

var (
	container = startContainer()
)

type ClientTestSuite struct {
	suite.Suite
	client Client
}

func (suite *ClientTestSuite) SetupTest() {
	client := createClient(container)
	err := client.Authenticate()
	require.NoError(suite.T(), err)
	suite.client = client
}

func TestClientIntegration(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestClient_Authenticate() {
	err := suite.client.Authenticate()

	assert.NoError(suite.T(), err)
}

func (suite *ClientTestSuite) TestClient_DeleteSession() {
	err := suite.client.DeleteSession()

	assert.NoError(suite.T(), err)
}

func (suite *ClientTestSuite) TestClient_GetVersion() {
	version, err := suite.client.GetVersion()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), version)
}

func (suite *ClientTestSuite) TestClient_GetTeleporter() {
	payload, err := suite.client.GetTeleporter()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), payload)
}

func (suite *ClientTestSuite) TestClient_PostTeleporter() {
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

func (suite *ClientTestSuite) TestClient_GetConfig() {
	conf, err := suite.client.GetConfig()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), conf)
}

func (suite *ClientTestSuite) TestClient_PatchConfig() {
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
	assert.Error(t, a.verify())

	a.sid = "sid123"
	assert.Error(t, a.verify())

	a.validity = 1
	assert.NoError(t, a.verify())

}

func startContainer() testcontainers.Container {
	containerRequest := testcontainers.ContainerRequest{
		Image:        dockerImage,
		ExposedPorts: []string{"80/tcp", "53/tcp", "53/udp"},
		WaitingFor:   wait.ForListeningPort("80").WithStartupTimeout(30 * time.Second),
		Env: map[string]string{
			"FTLCONF_dns_upstreams":          "8.8.8.8",
			"FTLCONF_webserver_api_password": apiPassword,
		},
	}

	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})

	if err != nil {
		log.Fatalf("starting pihole test container: %v", err)
	}
	return container
}

func createClient(container testcontainers.Container) Client {
	apiPort, err := container.MappedPort(context.Background(), "80/tcp")
	if err != nil {
		panic(err)
	}

	host := fmt.Sprintf("http://localhost:%s", apiPort.Port())

	return NewClient(model.NewPiHole(host, "test"))
}
