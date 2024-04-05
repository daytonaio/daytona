// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package frpc

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/server/config"
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
