// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

// ContainerRegistry represents a container registry credentials
type ContainerRegistry struct {
	Server   string `json:"server" validate:"required" gorm:"primaryKey"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
} // @name ContainerRegistry
