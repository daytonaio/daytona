// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	log "github.com/sirupsen/logrus"
)

const registryContainerName = "daytona-registry"

type LocalContainerRegistryConfig struct {
	DataPath          string
	Port              uint32
	Image             string
	ContainerRegistry *models.ContainerRegistry
	Logger            io.Writer
	Frps              *server.FRPSConfig
	ServerId          string
}

func NewLocalContainerRegistry(config *LocalContainerRegistryConfig) *LocalContainerRegistry {
	return &LocalContainerRegistry{
		dataPath:          config.DataPath,
		port:              config.Port,
		image:             config.Image,
		containerRegistry: config.ContainerRegistry,
		logger:            config.Logger,
		frps:              config.Frps,
		serverId:          config.ServerId,
	}
}

type LocalContainerRegistry struct {
	dataPath          string
	port              uint32
	image             string
	containerRegistry *models.ContainerRegistry
	logger            io.Writer
	frps              *server.FRPSConfig
	serverId          string
}

func (s *LocalContainerRegistry) Start() error {
	ctx := context.Background()

	_, err := os.Stat(s.dataPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(s.dataPath, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if _, err := cli.Info(ctx); err != nil {
		return errors.New("cannot connect to the Docker daemon. Is the Docker daemon running?\nIf Docker is not installed, please install it by following https://docs.docker.com/engine/install/ and try again")
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	//	we want to always create a new container
	//	to avoid conflicts with configuration changes
	if err := RemoveRegistryContainer(); err != nil {
		return err
	}

	_, err = net.Dial("tcp", fmt.Sprintf(":%d", s.port))
	if err == nil {
		return fmt.Errorf("cannot start registry, port %d is already in use", s.port)
	}

	// Pull the image
	err = dockerClient.PullImage(s.image, s.containerRegistry, s.logger)
	if err != nil {
		return err
	}

	//	todo: enable TLS
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: s.image,
		Env: []string{
			fmt.Sprintf("REGISTRY_HTTP_ADDR=0.0.0.0:%d", s.port),
		},
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", s.port)): {},
		},
	}, &container.HostConfig{
		Privileged: true,
		Binds: []string{
			s.dataPath + ":/var/lib/registry",
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", s.port)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprint(s.port),
				},
			},
		},
	}, nil, nil, registryContainerName)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		errChan <- cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	}()

	if s.frps == nil {
		return <-errChan
	}

	healthCheck, frpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: s.frps.Domain,
		ServerPort:   int(s.frps.Port),
		Name:         fmt.Sprintf("daytona-server-registry-%s", s.serverId),
		Port:         int(s.port),
		SubDomain:    fmt.Sprintf("registry-%s", s.serverId),
	})
	if err != nil {
		return err
	}

	go func() {
		err := frpcService.Run(context.Background())
		if err != nil {
			errChan <- err
		}
	}()

	for i := 0; i < 5; i++ {
		if err = healthCheck(); err != nil {
			log.Debugf("Failed to connect to registry frpc: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return err
	}

	return <-errChan
}

func (s *LocalContainerRegistry) Stop() error {
	return RemoveRegistryContainer()
}

func (s *LocalContainerRegistry) Purge() error {
	err := s.Stop()
	if err != nil {
		return err
	}

	return os.RemoveAll(s.dataPath)
}

func RemoveRegistryContainer() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	ctx := context.Background()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == fmt.Sprintf("/%s", registryContainerName) {
				removeOptions := container.RemoveOptions{
					Force: true,
				}

				if err := cli.ContainerRemove(ctx, c.ID, removeOptions); err != nil {
					return fmt.Errorf("failed to remove local container registry: %w", err)
				}
				return nil
			}
		}
	}

	return nil
}
