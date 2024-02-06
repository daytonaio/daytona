// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	workspace_proto "github.com/daytonaio/daytona/grpc/proto"
)

type WorkspaceServer struct {
	workspace_proto.UnimplementedWorkspaceServer
}
