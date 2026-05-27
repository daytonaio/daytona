// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// Typed runner errors. Each implements common_errors.HTTPError; producers
// return them directly and the middleware writes the envelope as-is.

type StorageExpansionLimitReachedError struct{ Message string }

func NewStorageExpansionLimitReachedError(message string) *StorageExpansionLimitReachedError {
	return &StorageExpansionLimitReachedError{Message: message}
}
func (e *StorageExpansionLimitReachedError) Error() string       { return e.Message }
func (e *StorageExpansionLimitReachedError) HTTPStatusCode() int { return http.StatusConflict }
func (e *StorageExpansionLimitReachedError) ErrorCode() string {
	return string(CodeStorageExpansionLimitReached)
}

type DockerDaemonUnreachableError struct{ Message string }

func NewDockerDaemonUnreachableError(message string) *DockerDaemonUnreachableError {
	return &DockerDaemonUnreachableError{Message: message}
}
func (e *DockerDaemonUnreachableError) Error() string       { return e.Message }
func (e *DockerDaemonUnreachableError) HTTPStatusCode() int { return http.StatusBadGateway }
func (e *DockerDaemonUnreachableError) ErrorCode() string {
	return string(CodeDockerDaemonUnreachable)
}

type SandboxDaemonUnreachableError struct{ Message string }

func NewSandboxDaemonUnreachableError(message string) *SandboxDaemonUnreachableError {
	return &SandboxDaemonUnreachableError{Message: message}
}
func (e *SandboxDaemonUnreachableError) Error() string       { return e.Message }
func (e *SandboxDaemonUnreachableError) HTTPStatusCode() int { return http.StatusBadGateway }
func (e *SandboxDaemonUnreachableError) ErrorCode() string {
	return string(CodeSandboxDaemonUnreachable)
}

type SnapshotPullTimeoutError struct{ Message string }

func NewSnapshotPullTimeoutError(message string) *SnapshotPullTimeoutError {
	return &SnapshotPullTimeoutError{Message: message}
}
func (e *SnapshotPullTimeoutError) Error() string       { return e.Message }
func (e *SnapshotPullTimeoutError) HTTPStatusCode() int { return http.StatusGatewayTimeout }
func (e *SnapshotPullTimeoutError) ErrorCode() string   { return string(CodeSnapshotPullTimeout) }

type BuildTimeoutError struct{ Message string }

func NewBuildTimeoutError(message string) *BuildTimeoutError {
	return &BuildTimeoutError{Message: message}
}
func (e *BuildTimeoutError) Error() string       { return e.Message }
func (e *BuildTimeoutError) HTTPStatusCode() int { return http.StatusGatewayTimeout }
func (e *BuildTimeoutError) ErrorCode() string   { return string(CodeBuildTimeout) }

// 502 (not 500): container failure is not a runner bug.
type ContainerCrashedError struct{ Message string }

func NewContainerCrashedError(message string) *ContainerCrashedError {
	return &ContainerCrashedError{Message: message}
}
func (e *ContainerCrashedError) Error() string       { return e.Message }
func (e *ContainerCrashedError) HTTPStatusCode() int { return http.StatusBadGateway }
func (e *ContainerCrashedError) ErrorCode() string   { return string(CodeContainerCrashed) }

// 422 (not 400): request is well-formed but unprocessable on this runner.
type UnsupportedArchitectureError struct{ Message string }

func NewUnsupportedArchitectureError(message string) *UnsupportedArchitectureError {
	return &UnsupportedArchitectureError{Message: message}
}
func (e *UnsupportedArchitectureError) Error() string       { return e.Message }
func (e *UnsupportedArchitectureError) HTTPStatusCode() int { return http.StatusUnprocessableEntity }
func (e *UnsupportedArchitectureError) ErrorCode() string {
	return string(CodeUnsupportedArchitecture)
}

// RecoverableError carries a pattern-matched transient failure in the legacy
// 400+JSON-blob shape consumed by the API's sanitizeSandboxError.
type RecoverableError struct{ Encoded string }

// NewRecoverableError returns nil when err doesn't match a recoverable pattern.
func NewRecoverableError(err error) *RecoverableError {
	encoded, ok := formatRecoverable(err)
	if !ok {
		return nil
	}
	return &RecoverableError{Encoded: encoded}
}
func (e *RecoverableError) Error() string       { return e.Encoded }
func (e *RecoverableError) HTTPStatusCode() int { return http.StatusBadRequest }
func (e *RecoverableError) ErrorCode() string   { return "BAD_REQUEST" }

var (
	_ common_errors.HTTPError = (*StorageExpansionLimitReachedError)(nil)
	_ common_errors.HTTPError = (*DockerDaemonUnreachableError)(nil)
	_ common_errors.HTTPError = (*SandboxDaemonUnreachableError)(nil)
	_ common_errors.HTTPError = (*SnapshotPullTimeoutError)(nil)
	_ common_errors.HTTPError = (*BuildTimeoutError)(nil)
	_ common_errors.HTTPError = (*ContainerCrashedError)(nil)
	_ common_errors.HTTPError = (*UnsupportedArchitectureError)(nil)
	_ common_errors.HTTPError = (*RecoverableError)(nil)
)
