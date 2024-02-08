// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"

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
