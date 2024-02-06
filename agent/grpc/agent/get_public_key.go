// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_grpc

import (
	"context"
	config_ssh_key "dagent/agent/config/ssh_key"
	daytona_proto "dagent/grpc/proto"
)

func (a *AgentServer) GetPublicKey(ctx context.Context, request *daytona_proto.GetPublicKeyRequest) (*daytona_proto.GetPublicKeyResponse, error) {
	publicKey, err := config_ssh_key.GetPublicKey()
	if err != nil {
		return nil, err
	}

	return &daytona_proto.GetPublicKeyResponse{
		PublicKey: publicKey,
	}, nil
}
