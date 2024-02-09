// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	config "github.com/daytonaio/daytona/server/config"
	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) SetConfig(ctx context.Context, request *types.ServerConfig) (*empty.Empty, error) {
	err := config.Save(request)
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
