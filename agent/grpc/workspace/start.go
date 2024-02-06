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

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Start(ctx context.Context, request *daytona_proto.WorkspaceStartRequest) (*empty.Empty, error) {
	w, err := workspace.LoadFromDB(request.Name)
	if err != nil {
		return nil, err
	}

	//	todo: workspace config should be read from the container labels
	//		  this is a temporary workaround to move fast
	credClient := &credentials.CredentialsClient{}

	extensions := []workspace.Extension{}

	vsc_server := vsc_server.VscServerExtension{}
	extensions = append(extensions, vsc_server)

	ssh := ssh.SshExtension{}
	extensions = append(extensions, ssh)

	w.Credentials = credClient
	w.Extensions = extensions

	if request.Project != "" {
		err = w.StartProject(request.Project)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		err = w.Start()
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return new(empty.Empty), nil
}
