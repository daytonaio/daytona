// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package frpc

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/types"
)

func ConnectServer() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	return frpc.Connect(frpc.FrpcConnectParams{
		ServerDomain: c.Frps.Domain,
		ServerPort:   int(c.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-%s", c.Id),
		Port:         int(c.HeadscalePort),
		SubDomain:    c.Id,
	})
}

func ConnectApi() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	return frpc.Connect(frpc.FrpcConnectParams{
		ServerDomain: c.Frps.Domain,
		ServerPort:   int(c.Frps.Port),
		Name:         fmt.Sprintf("daytona-server-api-%s", c.Id),
		Port:         int(c.ApiPort),
		SubDomain:    fmt.Sprintf("api-%s", c.Id),
	})
}

func GetApiDomain(c *types.ServerConfig) string {
	return fmt.Sprintf("api-%s", GetServerDomain(c))
}

func GetServerDomain(c *types.ServerConfig) string {
	return fmt.Sprintf("%s.%s", c.Id, c.Frps.Domain)
}

func GetServerUrl(c *types.ServerConfig) string {
	return fmt.Sprintf("%s://%s", c.Frps.Protocol, GetServerDomain(c))
}

func GetApiUrl(c *types.ServerConfig) string {
	return fmt.Sprintf("%s://%s", c.Frps.Protocol, GetApiDomain(c))
}
