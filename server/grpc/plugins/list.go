// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package plugins_grpc

import (
	"context"

	"github.com/daytonaio/daytona/common/grpc/proto"
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *PluginsServer) ListProvisionerPlugins(ctx context.Context, req *empty.Empty) (*proto.ProvisionerPluginList, error) {
	provisioners := provisioner_manager.GetProvisioners()
	pluginList := &proto.ProvisionerPluginList{}
	for _, provisioner := range provisioners {
		info, err := provisioner.GetInfo()
		if err != nil {
			return nil, err
		}

		pluginList.Plugins = append(pluginList.Plugins, &proto.ProvisionerPlugin{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	return pluginList, nil
}

func (s *PluginsServer) ListAgentServicePlugins(ctx context.Context, req *empty.Empty) (*proto.AgentServicePluginList, error) {
	agentServices := agent_service_manager.GetAgentServices()
	pluginList := &proto.AgentServicePluginList{}
	for _, agentService := range agentServices {
		info, err := agentService.GetInfo()
		if err != nil {
			return nil, err
		}

		pluginList.Plugins = append(pluginList.Plugins, &proto.AgentServicePlugin{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	return &proto.AgentServicePluginList{}, nil
}
