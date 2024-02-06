// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/provisioner"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/grpc/proto/types"

	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Info(ctx context.Context, request *daytona_proto.WorkspaceInfoRequest) (*types.WorkspaceInfo, error) {
	w, err := db.FindWorkspace(request.Id)
	if err != nil {
		return nil, err
	}

	log.Debug(w)

	return provisioner.GetWorkspaceInfo(w)
}
