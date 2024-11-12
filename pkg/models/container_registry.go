// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"errors"
	"strings"
)

// ContainerRegistry represents a container registry credentials
type ContainerRegistry struct {
	Server   string `json:"server" validate:"required" gorm:"primaryKey"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
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
