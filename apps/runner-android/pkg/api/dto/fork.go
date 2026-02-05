// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

// ForkSandboxDTO represents a request to fork a sandbox
type ForkSandboxDTO struct {
	NewSandboxId string `json:"newSandboxId" binding:"required" validate:"required"`
	Prefault     bool   `json:"prefault,omitempty"` // Prefault memory pages for faster access
} //	@name	ForkSandboxDTO

// ForkSandboxResponseDTO represents the response from a fork operation
type ForkSandboxResponseDTO struct {
	Id       string `json:"id"`
	State    string `json:"state"`
	ParentId string `json:"parentId"`
} //	@name	ForkSandboxResponseDTO
