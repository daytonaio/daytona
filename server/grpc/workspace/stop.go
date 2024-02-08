// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Stop(ctx context.Context, request *daytona_proto.WorkspaceStopRequest) (*empty.Empty, error) {
	w, err := db.FindWorkspace(request.Id)
	if err != nil {
		return nil, err
	}

	log.Debug(w)

	if request.Project != "" {
		project, err := getProject(w, request.Project)
		if err != nil {
			return nil, err
		}

		err = provisioner.StopProject(project)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		err = provisioner.StopWorkspace(w)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return new(empty.Empty), nil
}
