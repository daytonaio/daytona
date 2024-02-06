// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"

	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/credentials"
	"github.com/daytonaio/daytona/extensions/ssh"
	"github.com/daytonaio/daytona/extensions/vsc_server"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Info(ctx context.Context, request *daytona_proto.WorkspaceInfoRequest) (*daytona_proto.WorkspaceInfoResponse, error) {
	w, err := workspace.LoadFromDB(request.Name)
	if err != nil {
		return nil, err
	}

	credClient := &credentials.CredentialsClient{}

	extensions := []workspace.Extension{}

	vsc_server := vsc_server.VscServerExtension{}
	extensions = append(extensions, vsc_server)

	ssh := ssh.SshExtension{}
	extensions = append(extensions, ssh)

	w.Credentials = credClient
	w.Extensions = extensions

	log.Debug(w)

	workspaceInfo, err := w.Info()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	projectInfos := []*daytona_proto.WorkspaceProjectInfo{}

	for _, projectInfo := range workspaceInfo.Projects {
		extensionInfos := []*daytona_proto.WorkspaceProjectExtensionInfo{}

		for _, extensionInfo := range projectInfo.Extensions {
			extensionInfos = append(extensionInfos, &daytona_proto.WorkspaceProjectExtensionInfo{
				Name: extensionInfo.Name,
				Info: extensionInfo.Info,
			})
		}

		running := false
		if projectInfo.ContainerInfo != nil {
			running = projectInfo.ContainerInfo.IsRunning
		}

		projectInfos = append(projectInfos, &daytona_proto.WorkspaceProjectInfo{
			Name:       projectInfo.Name,
			Available:  projectInfo.Available,
			Running:    running,
			Extensions: extensionInfos,
		})
	}

	return &daytona_proto.WorkspaceInfoResponse{
		Name:     workspaceInfo.Name,
		Cwd:      workspaceInfo.Cwd,
		Projects: projectInfos,
	}, nil
}
