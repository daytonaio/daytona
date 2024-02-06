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

func (m *WorkspaceServer) Stop(ctx context.Context, request *daytona_proto.WorkspaceStopRequest) (*empty.Empty, error) {
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

	log.Debug(w)

	if request.Project != "" {
		err = w.StopProject(request.Project)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		err = w.Stop()
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return new(empty.Empty), nil
}
