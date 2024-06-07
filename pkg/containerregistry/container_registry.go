// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"errors"
	"strings"
)

// ContainerRegistry represents a container registry credentials
type ContainerRegistry struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
} // @name ContainerRegistry

func GetServerHostname(server string) (string, error) {
	server = strings.TrimPrefix(server, "https://")
	server = strings.TrimPrefix(server, "http://")

	parts := strings.Split(server, "/")

	if len(parts) == 0 {
		return "", errors.New("invalid container registry server URL")
	}

	return parts[0], nil
}
