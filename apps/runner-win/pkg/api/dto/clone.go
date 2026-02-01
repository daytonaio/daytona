// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

// CloneSandboxDTO represents a request to clone a sandbox
type CloneSandboxDTO struct {
	NewSandboxId string `json:"newSandboxId" binding:"required"`
} //	@name	CloneSandboxDTO

// CloneSandboxResponseDTO represents the response from a clone operation
type CloneSandboxResponseDTO struct {
	Id              string `json:"id"`
	State           string `json:"state"`
	SourceSandboxId string `json:"sourceSandboxId"`
	DaemonVersion   string `json:"daemonVersion,omitempty"`
} //	@name	CloneSandboxResponseDTO
