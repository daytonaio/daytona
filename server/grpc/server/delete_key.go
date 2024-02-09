// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	config_ssh_key "github.com/daytonaio/daytona/server/config/ssh_key"

	"github.com/golang/protobuf/ptypes/empty"
)

func (a *ServerGRPCServer) DeleteKey(ctx context.Context, request *empty.Empty) (*empty.Empty, error) {
	err := config_ssh_key.DeletePrivateKey()
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
