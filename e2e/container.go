package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const startupTimeout = 30 * time.Second

type PiHoleContainer struct {
	Container tc.Container
	password  string
}

func (c *PiHoleContainer) ConnectionString(ssl bool) string {
	var protocol string
	var port string

	if ssl {
		protocol = "https"
		port = "443/tcp"
	} else {
		protocol = "http"
		port = "80/tcp"
	}

	mappedPort, err := c.Container.MappedPort(context.Background(), nat.Port(port))
	if err != nil {
		panic(err)
	}

	hostIP, err := c.Container.Host(context.Background())
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s://%s:%s", protocol, hostIP, mappedPort.Port())
}

func (c *PiHoleContainer) EnvString(ssl bool) string {
	return fmt.Sprintf("%s|%s", c.ConnectionString(ssl), c.password)
}

func RunPiHole(password string) *PiHoleContainer {
	logStrategy := wait.ForLog("listening on")
	portStrategy := wait.ForListeningPort("80").WithStartupTimeout(startupTimeout)

	containerReq := tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "pihole/pihole:latest",
			ExposedPorts: []string{"80/tcp", "443/tcp"},
			WaitingFor:   wait.ForAll(portStrategy, logStrategy),
			Env: map[string]string{
				"FTLCONF_webserver_api_password": password,
			},
		},
		Started: true,
	}

	container, err := tc.GenericContainer(context.Background(), containerReq)
	if err != nil {
		panic(err)
	}

	return &PiHoleContainer{
		Container: container,
		password:  password,
	}
}
