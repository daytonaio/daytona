// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	config "github.com/daytonaio/daytona/server/config"
	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) GetConfig(ctx context.Context, request *empty.Empty) (*types.ServerConfig, error) {
	return config.GetConfig()
}
