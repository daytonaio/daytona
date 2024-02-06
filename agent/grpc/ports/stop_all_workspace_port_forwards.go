// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/port_manager"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
)

func (p *PortsServer) StopAllWorkspacePortForwards(ctx context.Context, request *daytona_proto.StopAllWorkspacePortForwardsRequest) (*empty.Empty, error) {
	w, err := db.FindWorkspace(request.WorkspaceId)
	if err != nil {
		return nil, err
	}

	err = port_manager.StopAllWorkspaceForwards(w.Name)
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
