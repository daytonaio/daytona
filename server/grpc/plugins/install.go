// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package plugins_grpc

import (
	"context"
	"path"

	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/os"
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	"github.com/daytonaio/daytona/plugins/plugin_manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/frpc"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *PluginsServer) InstallProvisionerPlugin(ctx context.Context, req *proto.InstallPluginRequest) (*empty.Empty, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	downloadPath := path.Join(c.PluginsDir, "provisioners", req.Name, req.Name)

	err = plugin_manager.DownloadPlugin(convertDownloadUrls(req.DownloadUrls), downloadPath)
	if err != nil {
		return nil, err
	}

	err = provisioner_manager.RegisterProvisioner(downloadPath, c.ServerDownloadUrl, frpc.GetServerUrl(c), frpc.GetApiUrl(c))
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *PluginsServer) InstallAgentServicePlugin(ctx context.Context, req *proto.InstallPluginRequest) (*empty.Empty, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	downloadPath := path.Join(c.PluginsDir, "agent_services", req.Name, req.Name)

	err = plugin_manager.DownloadPlugin(convertDownloadUrls(req.DownloadUrls), downloadPath)
	if err != nil {
		return nil, err
	}

	err = agent_service_manager.RegisterAgentService(downloadPath)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func convertDownloadUrls(downloadUrls map[string]string) map[os.OperatingSystem]string {
	converted := make(map[os.OperatingSystem]string)
	for k, v := range downloadUrls {
		converted[os.OperatingSystem(k)] = v
	}
	return converted
}
