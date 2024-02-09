// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	config "github.com/daytonaio/daytona/server/config"
	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) GetConfig(ctx context.Context, request *empty.Empty) (*daytona_proto.GetConfigResponse, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	return &daytona_proto.GetConfigResponse{
		DefaultWorkspaceDir: config.DefaultWorkspaceDir,
		ProjectBaseImage:    config.ProjectBaseImage,
		PluginsDir:          config.PluginsDir,
	}, nil
}
