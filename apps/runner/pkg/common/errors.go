// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package common holds the runner's error vocabulary (see runner_errors.go).
package common

import (
	"time"
)

// RunnerErrorCode is the machine-readable error code surfaced to SDK callers.
type RunnerErrorCode string // @name RunnerErrorCode

const (
	CodeStorageExpansionLimitReached RunnerErrorCode = "STORAGE_EXPANSION_LIMIT_REACHED"
	CodeDockerDaemonUnreachable      RunnerErrorCode = "DOCKER_DAEMON_UNREACHABLE"
	CodeSandboxDaemonUnreachable     RunnerErrorCode = "SANDBOX_DAEMON_UNREACHABLE"
	CodeSnapshotPullTimeout          RunnerErrorCode = "SNAPSHOT_PULL_TIMEOUT"
	CodeBuildTimeout                 RunnerErrorCode = "BUILD_TIMEOUT"
	CodeContainerCrashed             RunnerErrorCode = "CONTAINER_CRASHED"
	CodeUnsupportedArchitecture      RunnerErrorCode = "UNSUPPORTED_ARCHITECTURE"
)

// ErrorResponse mirrors libs/common-go ErrorResponse; this runner-local copy
// exists only so swaggo emits a typed RunnerErrorCode enum reference.
//
//	@Description	Error response
//	@Schema			ErrorResponse
type ErrorResponse struct {
	StatusCode int             `json:"statusCode" example:"400" binding:"required"`
	Message    string          `json:"message" example:"Bad request" binding:"required"`
	Source     string          `json:"source,omitempty" example:"DAYTONA_RUNNER"`
	Code       RunnerErrorCode `json:"code,omitempty" example:"STORAGE_EXPANSION_LIMIT_REACHED"`
	Timestamp  time.Time       `json:"timestamp" example:"2023-01-01T12:00:00Z" binding:"required"`
	Path       string          `json:"path" example:"/api/resource" binding:"required"`
	Method     string          `json:"method,omitempty" example:"GET"`
} //	@name	ErrorResponse
