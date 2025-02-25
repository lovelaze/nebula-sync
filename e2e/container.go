package e2e

import (
	"context"
	"fmt"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

type PiHoleContainer struct {
	Container tc.Container
	password  string
}

func (c *PiHoleContainer) ConnectionString() string {
	mappedPort, err := c.Container.MappedPort(context.Background(), "80/tcp")
	if err != nil {
		panic(err)
	}

	hostIP, err := c.Container.Host(context.Background())
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())
}

func (c *PiHoleContainer) EnvString() string {
	return fmt.Sprintf("%s|%s", c.ConnectionString(), c.password)
}

func RunPiHole(password string) *PiHoleContainer {
	logStrategy := wait.ForLog("listening on")
	portStrategy := wait.ForListeningPort("80").WithStartupTimeout(30 * time.Second)

	containerReq := tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "pihole/pihole:latest",
			ExposedPorts: []string{"80/tcp"},
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
