// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"
	"dagent/agent/port_manager"
	"dagent/agent/workspace"
	daytona_proto "dagent/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
)

func (p *PortsServer) StopPortForward(ctx context.Context, request *daytona_proto.StopPortForwardRequest) (*empty.Empty, error) {
	w, err := workspace.LoadFromDB(request.WorkspaceName)
	if err != nil {
		return nil, err
	}

	project, err := w.GetProject(request.Project)
	if err != nil {
		return nil, err
	}

	containerName := project.GetContainerName()

	err = port_manager.StopPortForward(w.Name, containerName, port_manager.ContainerPort(request.Port))
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
