// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/server/headscale"
	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) GenerateAuthKey(ctx context.Context, request *empty.Empty) (*daytona_proto.AuthKey, error) {
	authKey, err := headscale.CreateAuthKey()
	if err != nil {
		return nil, err
	}

	return &daytona_proto.AuthKey{
		Key: authKey,
	}, nil
}
