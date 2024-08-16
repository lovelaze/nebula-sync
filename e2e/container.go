package e2e

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

type PiHoleContainer struct {
	Container testcontainers.Container
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

func RunPiHole(ctx context.Context, password string, opts ...testcontainers.ContainerCustomizer) *PiHoleContainer {
	req := testcontainers.ContainerRequest{
		Image:        "pihole/pihole:development-v6",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80").WithStartupTimeout(10 * time.Second),
		Env: map[string]string{
			"FTLCONF_webserver_api_password": password,
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			panic(err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		panic(err)
	}

	return &PiHoleContainer{Container: container, password: password}
}
