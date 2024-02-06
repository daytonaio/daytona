// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"context"
	"dagent/agent/workspace"
	"dagent/credentials"
	"dagent/extensions/ssh"
	"dagent/extensions/vsc_server"
	daytona_proto "dagent/grpc/proto"

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

		response = append(response, &daytona_proto.WorkspaceInfoResponse{
			Name:     workspaceInfo.Name,
			Cwd:      workspaceInfo.Cwd,
			Projects: projectInfos,
		})
	}

	return &daytona_proto.WorkspaceListResponse{
		Workspaces: response,
	}, nil
}
