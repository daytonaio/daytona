// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

// ContainerRegistry represents a container registry credentials
type ContainerRegistry struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
} // @name ContainerRegistry
