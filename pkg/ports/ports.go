// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/daytonaio/daytona/internal/tailscale"
	log "github.com/sirupsen/logrus"
	"tailscale.com/tsnet"
)

func GetAvailableEphemeralPort() (uint16, error) {
	var ephemeralPort uint16
	for ephemeralPort = 50000; ephemeralPort < 60000; ephemeralPort++ {
		if IsPortAvailable(ephemeralPort) {
			log.Debug("EPHEMERAL PORT: " + strconv.FormatUint(uint64(ephemeralPort), 10))
			return ephemeralPort, nil
		}
	}
	return 0, errors.New("no more ephemeral ports available")
}

func IsPortAvailable(port uint16) bool {
	timeout := time.Millisecond * 50
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return true
	}
	if conn != nil {
		defer conn.Close()
	}
	return false
}

func ForwardPort(workspaceId, projectName string, targetPort uint16) (*uint16, chan error) {
	hostPort := targetPort
	errChan := make(chan error)
	var err error
	if !IsPortAvailable(targetPort) {
		hostPort, err = GetAvailableEphemeralPort()
		if err != nil {
			errChan <- err
			return nil, errChan
		}
	}

	tsConn, err := tailscale.GetConnection(nil)
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

			targetUrl := fmt.Sprintf("%s-%s:%d", workspaceId, projectName, targetPort)

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
