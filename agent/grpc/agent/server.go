// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_grpc

import (
	agent_proto "dagent/grpc/proto"
)

type AgentServer struct {
	agent_proto.UnimplementedAgentServer
}
