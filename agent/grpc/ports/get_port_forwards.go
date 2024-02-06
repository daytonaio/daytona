// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"
	"dagent/agent/port_manager"
	"dagent/agent/workspace"
	daytona_proto "dagent/grpc/proto"
)

func (p *PortsServer) GetPortForwards(ctx context.Context, request *daytona_proto.GetPortForwardsRequest) (*daytona_proto.WorkspacePortForward, error) {
	w, err := workspace.LoadFromDB(request.WorkspaceName)
	if err != nil {
		return nil, err
	}

	portForwards, err := port_manager.GetPortForwards(w.Name)
	if err != nil {
		return nil, err
	}

	projectPortForwards := make(map[string]*daytona_proto.ProjectPortForwards)

	for project, portForward := range portForwards.ProjectPortForwards {
		projectPortForward := &daytona_proto.ProjectPortForwards{
			Project:      project,
			PortForwards: make(map[uint32]*daytona_proto.PortForward),
		}

		for containerPort, portForward := range portForward {
			projectPortForward.PortForwards[uint32(containerPort)] = &daytona_proto.PortForward{
				ContainerPort: uint32(portForward.ContainerPort),
				HostPort:      uint32(portForward.HostPort),
			}
		}

		projectPortForwards[project] = projectPortForward
	}

	return &daytona_proto.WorkspacePortForward{
		WorkspaceName:       w.Name,
		ProjectPortForwards: projectPortForwards,
	}, nil
}
