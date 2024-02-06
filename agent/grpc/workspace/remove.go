// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/provisioner"
	"github.com/daytonaio/daytona/agent/workspace"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Remove(ctx context.Context, request *daytona_proto.WorkspaceRemoveRequest) (*empty.Empty, error) {
	w, err := workspace.LoadFromDB(request.Name)
	if err != nil {
		return nil, err
	}

	log.Debug(w)

	err = provisioner.DestroyWorkspace(*w)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = workspace.DeleteFromDB(w)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return new(empty.Empty), nil
}
