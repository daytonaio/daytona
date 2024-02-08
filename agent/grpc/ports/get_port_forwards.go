// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/port_manager"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
)

func (p *PortsServer) GetPortForwards(ctx context.Context, request *daytona_proto.GetPortForwardsRequest) (*daytona_proto.WorkspacePortForward, error) {
	w, err := db.FindWorkspace(request.WorkspaceId)
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
		WorkspaceId:         w.Id,
		ProjectPortForwards: projectPortForwards,
	}, nil
}
