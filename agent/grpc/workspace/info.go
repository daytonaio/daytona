// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/agent/provisioner"
	"github.com/daytonaio/daytona/agent/workspace"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Info(ctx context.Context, request *daytona_proto.WorkspaceInfoRequest) (*daytona_proto.WorkspaceInfoResponse, error) {
	w, err := workspace.LoadFromDB(request.Name)
	if err != nil {
		return nil, err
	}

	log.Debug(w)

	return getWorkspaceInfo(*w)
}

func getWorkspaceInfo(w workspace.Workspace) (*daytona_proto.WorkspaceInfoResponse, error) {
	workspaceInfo, err := provisioner.GetWorkspaceInfo(w)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	workspaceMetadata, err := json.Marshal(workspaceInfo.ProvisionerMetadata)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	projectInfos := []*daytona_proto.WorkspaceProjectInfo{}

	for _, projectInfo := range workspaceInfo.Projects {
		metadata, err := json.Marshal(projectInfo.ProvisionerMetadata)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		projectInfos = append(projectInfos, &daytona_proto.WorkspaceProjectInfo{
			Name:                projectInfo.Name,
			Created:             projectInfo.Created,
			Started:             projectInfo.Started,
			Finished:            projectInfo.Finished,
			IsRunning:           projectInfo.IsRunning,
			ProvisionerMetadata: string(metadata),
		})
	}

	return &daytona_proto.WorkspaceInfoResponse{
		Name:     workspaceInfo.Name,
		Projects: projectInfos,
		Provisioner: &daytona_proto.WorkspaceProvisioner{
			Name:    workspaceInfo.Provisioner.Name,
			Profile: workspaceInfo.Provisioner.Profile,
		},
		ProvisionerMetadata: string(workspaceMetadata),
	}, nil
}
