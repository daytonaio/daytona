// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/server"
)

func ToServerConfig(serverConfigDto *apiclient.ServerConfig) *server.Config {
	if serverConfigDto == nil {
		return nil
	}

	config := &server.Config{
		Id:                              *serverConfigDto.Id,
		ProvidersDir:                    *serverConfigDto.ProvidersDir,
		RegistryUrl:                     *serverConfigDto.RegistryUrl,
		ServerDownloadUrl:               *serverConfigDto.ServerDownloadUrl,
		IpWithProtocol:                  serverConfigDto.IpWithProtocol,
		ApiPort:                         uint32(*serverConfigDto.ApiPort),
		LocalBuilderRegistryPort:        uint32(*serverConfigDto.LocalBuilderRegistryPort),
		BuilderRegistryServer:           *serverConfigDto.BuilderRegistryServer,
		BuildImageNamespace:             *serverConfigDto.BuildImageNamespace,
		HeadscalePort:                   uint32(*serverConfigDto.HeadscalePort),
		BinariesPath:                    *serverConfigDto.BinariesPath,
		LogFilePath:                     *serverConfigDto.LogFilePath,
		DefaultProjectImage:             *serverConfigDto.DefaultProjectImage,
		DefaultProjectUser:              *serverConfigDto.DefaultProjectUser,
		DefaultProjectPostStartCommands: serverConfigDto.DefaultProjectPostStartCommands,
		BuilderImage:                    *serverConfigDto.BuilderImage,
	}

	if serverConfigDto.Frps != nil {
		config.Frps = &server.FRPSConfig{
			Domain:   *serverConfigDto.Frps.Domain,
			Port:     uint32(*serverConfigDto.Frps.Port),
			Protocol: *serverConfigDto.Frps.Protocol,
		}
	}

	return config
}
