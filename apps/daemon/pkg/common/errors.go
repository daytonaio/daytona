// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"time"
)

// DaemonErrorCode identifies the specific error within the daemon.
type DaemonErrorCode string // @name DaemonErrorCode

const (
	// Generic
	CodeBadRequest          DaemonErrorCode = "BAD_REQUEST"
	CodeUnauthorized        DaemonErrorCode = "UNAUTHORIZED"
	CodeForbidden           DaemonErrorCode = "FORBIDDEN"
	CodeNotFound            DaemonErrorCode = "NOT_FOUND"
	CodeConflict            DaemonErrorCode = "CONFLICT"
	CodeInvalidRequestBody  DaemonErrorCode = "INVALID_REQUEST_BODY"
	CodeInternalServerError DaemonErrorCode = "INTERNAL_SERVER_ERROR"

	// Git
	CodeGitAuthFailed    DaemonErrorCode = "GIT_AUTH_FAILED"
	CodeGitAuthForbidden DaemonErrorCode = "GIT_AUTH_FORBIDDEN"
	CodeGitRepoNotFound  DaemonErrorCode = "GIT_REPO_NOT_FOUND"
	CodeGitBranchNotFound DaemonErrorCode = "GIT_BRANCH_NOT_FOUND"
	CodeGitRefNotFound   DaemonErrorCode = "GIT_REF_NOT_FOUND"
	CodeGitEmptyRepo     DaemonErrorCode = "GIT_EMPTY_REPO"
	CodeGitPushRejected  DaemonErrorCode = "GIT_PUSH_REJECTED"
	CodeGitDirtyWorktree DaemonErrorCode = "GIT_DIRTY_WORKTREE"
	CodeGitBranchExists  DaemonErrorCode = "GIT_BRANCH_EXISTS"
	CodeGitMergeConflict DaemonErrorCode = "GIT_MERGE_CONFLICT"
	CodeGitRepoExists    DaemonErrorCode = "GIT_REPO_EXISTS"

	// File system
	CodeFileNotFound    DaemonErrorCode = "FILE_NOT_FOUND"
	CodeFileAccessDenied DaemonErrorCode = "FILE_ACCESS_DENIED"
	CodeInvalidFilePath DaemonErrorCode = "INVALID_FILE_PATH"
	CodeFileReadFailed  DaemonErrorCode = "FILE_READ_FAILED"
)

// ErrorResponse represents the error response structure
//
//	@Description	Error response
//	@Schema			ErrorResponse
type ErrorResponse struct {
	StatusCode int             `json:"statusCode" example:"400" binding:"required"`
	Message    string          `json:"message" example:"Bad request" binding:"required"`
	Source     string          `json:"source" example:"DAYTONA_DAEMON" binding:"required"`
	Code       DaemonErrorCode `json:"code" example:"GIT_REPO_NOT_FOUND" binding:"required"`
	Timestamp  time.Time       `json:"timestamp" example:"2023-01-01T12:00:00Z" binding:"required"`
	Path       string          `json:"path" example:"/api/resource" binding:"required"`
	Method     string          `json:"method" example:"GET" binding:"required"`
} //	@name	ErrorResponse
