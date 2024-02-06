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

func (m *WorkspaceServer) Start(ctx context.Context, request *daytona_proto.WorkspaceStartRequest) (*empty.Empty, error) {
	w, err := workspace.LoadFromDB(request.Name)
	if err != nil {
		return nil, err
	}

	if request.Project != "" {
		project, err := w.GetProject(request.Project)
		if err != nil {
			return nil, err
		}

		err = provisioner.StartProject(*project)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		err = provisioner.StartWorkspace(*w)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return new(empty.Empty), nil
}
