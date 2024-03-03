// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package frpc

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func ConnectServer() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	cfg := client.ServiceOptions{}
	cfg.Common = &v1.ClientCommonConfig{}
	cfg.Common.ServerAddr = c.Frps.Domain
	cfg.Common.ServerPort = int(c.Frps.Port)
	cfg.ProxyCfgs = []v1.ProxyConfigurer{}

	httpConfig := &v1.HTTPProxyConfig{}
	httpConfig.GetBaseConfig().Name = fmt.Sprintf("daytona-server-%s", c.Id)
	httpConfig.GetBaseConfig().LocalPort = int(c.HeadscalePort)
	httpConfig.GetBaseConfig().Type = string(v1.ProxyTypeHTTP)
	httpConfig.SubDomain = c.Id

	cfg.ProxyCfgs = append(cfg.ProxyCfgs, httpConfig)

	service, err := client.NewService(cfg)
	if err != nil {
		return err
	}

	return service.Run(context.Background())
}
func ConnectApi() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	cfg := client.ServiceOptions{}
	cfg.Common = &v1.ClientCommonConfig{}
	cfg.Common.ServerAddr = c.Frps.Domain
	cfg.Common.ServerPort = int(c.Frps.Port)
	cfg.ProxyCfgs = []v1.ProxyConfigurer{}

	httpConfig := &v1.HTTPProxyConfig{}
	httpConfig.GetBaseConfig().Name = fmt.Sprintf("daytona-server-api-%s", c.Id)
	httpConfig.GetBaseConfig().LocalPort = int(c.ApiPort)
	httpConfig.GetBaseConfig().Type = string(v1.ProxyTypeHTTP)
	httpConfig.SubDomain = fmt.Sprintf("api-%s", c.Id)

	cfg.ProxyCfgs = append(cfg.ProxyCfgs, httpConfig)

	service, err := client.NewService(cfg)
	if err != nil {
		return err
	}

	return service.Run(context.Background())
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
