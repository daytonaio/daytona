// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/provisioner"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/grpc/proto/types"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) List(ctx context.Context, request *empty.Empty) (*daytona_proto.WorkspaceListResponse, error) {
	workspaces, err := db.ListWorkspaces()
	if err != nil {
		return nil, err
	}

	workspaceInfos := []*types.WorkspaceInfo{}

	for _, workspace := range workspaces {
		workspaceInfo, err := provisioner.GetWorkspaceInfo(workspace)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		workspaceInfos = append(workspaceInfos, workspaceInfo)
	}

	return &daytona_proto.WorkspaceListResponse{
		Workspaces: workspaceInfos,
	}, nil
}
