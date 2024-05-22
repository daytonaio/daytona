// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type LocalContainerRegistryConfig struct {
	DataPath string
	Port     uint32
}

func NewLocalContainerRegistry(config *LocalContainerRegistryConfig) *LocalContainerRegistry {
	return &LocalContainerRegistry{
		dataPath: config.DataPath,
		port:     config.Port,
	}
}

type LocalContainerRegistry struct {
	dataPath string
	port     uint32
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

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/daytona-registry" {
				//	we want to always create a new container
				//	to avoid conflicts with configuration changes
				removeOptions := container.RemoveOptions{
					Force: true,
				}

				if err := cli.ContainerRemove(ctx, c.ID, removeOptions); err != nil {
					return err
				}
				break
			}
		}
	}

	portStr := strconv.FormatUint(uint64(s.port), 10)
	port, err := nat.NewPort("tcp", portStr)
	if err != nil {
		return err
	}

	// Pull the image
	out, err := cli.ImagePull(ctx, "registry:2.8.3", image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	//	todo: log progress

	// Wait for the image pull to complete
	_, err = io.Copy(io.Discard, out)
	if err != nil {
		return err
	}

	//	todo: enable TLS
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "registry:2.8.3",
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Env: []string{
			fmt.Sprintf("%s:%s", "REGISTRY_HTTP_ADDR=127.0.0.1", strconv.Itoa(int(s.port))),
		},
	}, &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		Binds: []string{
			s.dataPath + ":/var/lib/registry",
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
