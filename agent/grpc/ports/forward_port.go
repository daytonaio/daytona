// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports_grpc

import (
	"context"

	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
)

func (p *PortsServer) ForwardPort(ctx context.Context, request *daytona_proto.ForwardPortRequest) (*daytona_proto.PortForward, error) {
	panic("not implemented")
	// w, err := workspace.FindWorkspace(request.WorkspaceName)
	// if err != nil {
	// 	return nil, err
	// }

	// project, err := w.GetProject(request.Project)
	// if err != nil {
	// 	return nil, err
	// }

	// containerName := project.GetContainerName()

	// portForward, err := port_manager.ForwardPort(w.Name, containerName, port_manager.ContainerPort(request.Port))
	// if err != nil {
	// 	return nil, err
	// }

	// return &daytona_proto.PortForward{
	// 	ContainerPort: uint32(portForward.ContainerPort),
	// 	HostPort:      uint32(portForward.HostPort),
	// }, nil
}
