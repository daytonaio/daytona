// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_grpc

import (
	"context"

	config_ssh_key "github.com/daytonaio/daytona/agent/config/ssh_key"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
)

func (a *AgentServer) GenerateKey(ctx context.Context, request *daytona_proto.GenerateKeyRequest) (*daytona_proto.GetPublicKeyResponse, error) {
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
