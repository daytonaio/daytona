// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/provisioner"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/grpc/proto/types"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Start(ctx context.Context, request *daytona_proto.WorkspaceStartRequest) (*empty.Empty, error) {
	w, err := db.FindWorkspace(request.Id)
	if err != nil {
		return nil, err
	}

	if request.Project != "" {
		project, err := getProject(w, request.Project)
		if err != nil {
			return nil, err
		}

		err = provisioner.StartProject(project)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		err = provisioner.StartWorkspace(w)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return new(empty.Empty), nil
}

func getProject(workspace *types.Workspace, projectName string) (*types.Project, error) {
	for _, project := range workspace.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}
