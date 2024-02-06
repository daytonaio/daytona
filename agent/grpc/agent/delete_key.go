// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_grpc

import (
	"context"
	config_ssh_key "dagent/agent/config/ssh_key"
	daytona_proto "dagent/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
)

func (a *AgentServer) DeleteKey(ctx context.Context, request *daytona_proto.DeleteKeyRequest) (*empty.Empty, error) {
	err := config_ssh_key.DeletePrivateKey()
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
