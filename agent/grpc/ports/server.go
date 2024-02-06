// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	ports_proto "dagent/grpc/proto"
)

type PortsServer struct {
	ports_proto.UnimplementedPortsServer
}
