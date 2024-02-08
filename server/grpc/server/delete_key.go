// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	config_ssh_key "github.com/daytonaio/daytona/server/config/ssh_key"

	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) DeleteKey(ctx context.Context, request *daytona_proto.DeleteKeyRequest) (*empty.Empty, error) {
	err := config_ssh_key.DeletePrivateKey()
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
