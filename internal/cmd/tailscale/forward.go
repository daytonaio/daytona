// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"tailscale.com/tsnet"
)

func ForwardPort(workspaceId, projectName string, targetPort uint16, profile config.Profile) (*uint16, chan error) {
	hostPort := targetPort
	errChan := make(chan error)
	var err error
	if !ports.IsPortAvailable(targetPort) {
		hostPort, err = ports.GetAvailableEphemeralPort()
		if err != nil {
			errChan <- err
			return nil, errChan
		}
	}

	tsConn, err := GetConnection(&profile)
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	netListener, err := net.Listen("tcp", fmt.Sprintf(":%d", hostPort))
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	go func() {
		for {
			conn, err := netListener.Accept()
			if err != nil {
				errChan <- err
				return
			}

			targetUrl := fmt.Sprintf("%s:%d", project.GetProjectHostname(workspaceId, projectName), targetPort)

			go handlePortConnection(conn, tsConn, targetUrl, errChan)
		}
	}()

	return &hostPort, errChan
}

func handlePortConnection(conn net.Conn, tsConn *tsnet.Server, targetUrl string, errChan chan error) {
	dialConn, err := tsConn.Dial(context.Background(), "tcp", targetUrl)
	if err != nil {
		errChan <- err
		return
	}

	go func() {
		_, err := io.Copy(conn, dialConn)
		if err != nil {
			errChan <- err
		}
		conn.Close()
		dialConn.Close()
	}()

	go func() {
		_, err := io.Copy(dialConn, conn)
		if err != nil {
			errChan <- err
		}
		conn.Close()
		dialConn.Close()
	}()
}
