// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/port_manager"
	"github.com/daytonaio/daytona/agent/workspace"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
)

func (p *PortsServer) GetProjectPortForwards(ctx context.Context, request *daytona_proto.GetProjectPortForwardsRequest) (*daytona_proto.ProjectPortForwards, error) {
	w, err := workspace.LoadFromDB(request.WorkspaceName)
	if err != nil {
		return nil, err
	}

	project, err := w.GetProject(request.Project)
	if err != nil {
		return nil, err
	}

	portForwards, err := port_manager.GetProjectPortForwards(w.Name, project.GetContainerName())
	if err != nil {
		return nil, err
	}

	projectPortForward := &daytona_proto.ProjectPortForwards{
		Project:      request.Project,
		PortForwards: make(map[uint32]*daytona_proto.PortForward),
	}

	for containerPort, portForward := range portForwards {
		projectPortForward.PortForwards[uint32(containerPort)] = &daytona_proto.PortForward{
			ContainerPort: uint32(portForward.ContainerPort),
			HostPort:      uint32(portForward.HostPort),
		}
	}

	return projectPortForward, nil
}
