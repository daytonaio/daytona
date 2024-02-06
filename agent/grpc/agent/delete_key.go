// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_grpc

import (
	"context"

	config_ssh_key "github.com/daytonaio/daytona/agent/config/ssh_key"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
)

func (a *AgentServer) DeleteKey(ctx context.Context, request *daytona_proto.DeleteKeyRequest) (*empty.Empty, error) {
	err := config_ssh_key.DeletePrivateKey()
	if err != nil {
		return nil, err
	}

	return new(empty.Empty), nil
}
