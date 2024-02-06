// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/workspace"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) List(ctx context.Context, request *empty.Empty) (*daytona_proto.WorkspaceListResponse, error) {
	workspaces, err := workspace.ListFromDB()
	if err != nil {
		return nil, err
	}

	response := []*daytona_proto.WorkspaceInfoResponse{}

	for i, _ := range workspaces {
		w, err := workspace.LoadFromDB(workspaces[i].Name)
		if err != nil {
			return nil, err
		}

		log.Debug(w)

		workspaceInfo, err := getWorkspaceInfo(*w)
		if err != nil {
			return nil, err
		}

		response = append(response, workspaceInfo)
	}

	return &daytona_proto.WorkspaceListResponse{
		Workspaces: response,
	}, nil
}
