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

func (s *PluginsServer) UninstallProvisionerPlugin(ctx context.Context, req *proto.UninstallPluginRequest) (*empty.Empty, error) {
	err := provisioner_manager.UninstallProvisioner(req.Name)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *PluginsServer) UninstallAgentServicePlugin(ctx context.Context, req *proto.UninstallPluginRequest) (*empty.Empty, error) {
	err := agent_service_manager.UninstallAgentService(req.Name)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
