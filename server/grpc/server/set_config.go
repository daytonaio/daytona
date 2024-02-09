// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	config "github.com/daytonaio/daytona/server/config"
	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) SetConfig(ctx context.Context, request *daytona_proto.SetConfigRequest) (*empty.Empty, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	config.DefaultWorkspaceDir = request.DefaultWorkspaceDir
	config.ProjectBaseImage = request.ProjectBaseImage
	config.PluginsDir = request.PluginsDir

	err = config.Save()
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
