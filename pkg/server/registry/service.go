// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type LocalContainerRegistryConfig struct {
	DataPath string
	Port     uint32
	Image    string
}

func NewLocalContainerRegistry(config *LocalContainerRegistryConfig) *LocalContainerRegistry {
	return &LocalContainerRegistry{
		dataPath: config.DataPath,
		port:     config.Port,
		image:    config.Image,
	}
}

type LocalContainerRegistry struct {
	dataPath string
	port     uint32
	image    string
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

	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("cannot find Docker installation. Please install by following https://docs.docker.com/engine/install/ and try again")
	}

	if _, err := cli.Info(ctx); err != nil {
		return fmt.Errorf("cannot connect to the Docker daemon. Is the Docker daemon running?")
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
	err = dockerClient.PullImage(s.image, nil, os.Stdout)
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
	}, nil, nil, "daytona-registry")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
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
			if name == "/daytona-registry" {
				removeOptions := container.RemoveOptions{
					Force: true,
				}

				if err := cli.ContainerRemove(ctx, c.ID, removeOptions); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return nil
}
