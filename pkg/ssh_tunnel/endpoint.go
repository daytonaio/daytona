// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh_tunnel

import "fmt"

const (
	endpointTypeUnixSocket = "unix"
	endpointTypeTCP        = "tcp"
)

type Endpoint struct {
	host       string
	port       int
	unixSocket string
}

func (e *Endpoint) String() string {
	if e.unixSocket != "" {
		return e.unixSocket
	}
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

func (e *Endpoint) Type() string {
	if e.unixSocket != "" {
		return endpointTypeUnixSocket
	}
	return endpointTypeTCP
}

func NewTCPEndpoint(host string, port int) *Endpoint {
	return &Endpoint{
		host: host,
		port: port,
	}
}

func NewUnixEndpoint(socket string) *Endpoint {
	return &Endpoint{
		unixSocket: socket,
	}
}
