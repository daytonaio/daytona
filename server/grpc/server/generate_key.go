// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/common/grpc/proto"
	config_ssh_key "github.com/daytonaio/daytona/server/config/ssh_key"
)

func (a *ServerGRPCServer) GenerateKey(ctx context.Context, request *daytona_proto.GenerateKeyRequest) (*daytona_proto.GetPublicKeyResponse, error) {
	err := config_ssh_key.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	publicKey, err := config_ssh_key.GetPublicKey()
	if err != nil {
		return nil, err
	}

	return &daytona_proto.GetPublicKeyResponse{
		PublicKey: publicKey,
	}, nil
}
