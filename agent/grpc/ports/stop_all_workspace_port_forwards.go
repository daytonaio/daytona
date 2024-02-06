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

func (p *PortsServer) StopAllWorkspacePortForwards(ctx context.Context, request *daytona_proto.StopAllWorkspacePortForwardsRequest) (*empty.Empty, error) {
	w, err := workspace.LoadFromDB(request.WorkspaceName)
	if err != nil {
		return nil, err
	}

	err = port_manager.StopAllWorkspaceForwards(w.Name)
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
