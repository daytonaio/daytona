// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// Typed daemon errors. Each implements common_errors.HTTPError.

type GitAuthFailedError struct{ Message string }

func NewGitAuthFailedError(message string) *GitAuthFailedError {
	return &GitAuthFailedError{Message: message}
}
func (e *GitAuthFailedError) Error() string       { return e.Message }
func (e *GitAuthFailedError) HTTPStatusCode() int { return http.StatusUnauthorized }
func (e *GitAuthFailedError) ErrorCode() string   { return string(CodeGitAuthFailed) }

type GitRepoNotFoundError struct{ Message string }

func NewGitRepoNotFoundError(message string) *GitRepoNotFoundError {
	return &GitRepoNotFoundError{Message: message}
}
func (e *GitRepoNotFoundError) Error() string       { return e.Message }
func (e *GitRepoNotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *GitRepoNotFoundError) ErrorCode() string   { return string(CodeGitRepoNotFound) }

type GitBranchNotFoundError struct{ Message string }

func NewGitBranchNotFoundError(message string) *GitBranchNotFoundError {
	return &GitBranchNotFoundError{Message: message}
}
func (e *GitBranchNotFoundError) Error() string       { return e.Message }
func (e *GitBranchNotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *GitBranchNotFoundError) ErrorCode() string   { return string(CodeGitBranchNotFound) }

type GitBranchExistsError struct{ Message string }

func NewGitBranchExistsError(message string) *GitBranchExistsError {
	return &GitBranchExistsError{Message: message}
}
func (e *GitBranchExistsError) Error() string       { return e.Message }
func (e *GitBranchExistsError) HTTPStatusCode() int { return http.StatusConflict }
func (e *GitBranchExistsError) ErrorCode() string   { return string(CodeGitBranchExists) }

type GitPushRejectedError struct{ Message string }

func NewGitPushRejectedError(message string) *GitPushRejectedError {
	return &GitPushRejectedError{Message: message}
}
func (e *GitPushRejectedError) Error() string       { return e.Message }
func (e *GitPushRejectedError) HTTPStatusCode() int { return http.StatusConflict }
func (e *GitPushRejectedError) ErrorCode() string   { return string(CodeGitPushRejected) }

type GitDirtyWorktreeError struct{ Message string }

func NewGitDirtyWorktreeError(message string) *GitDirtyWorktreeError {
	return &GitDirtyWorktreeError{Message: message}
}
func (e *GitDirtyWorktreeError) Error() string       { return e.Message }
func (e *GitDirtyWorktreeError) HTTPStatusCode() int { return http.StatusConflict }
func (e *GitDirtyWorktreeError) ErrorCode() string   { return string(CodeGitDirtyWorktree) }

type GitMergeConflictError struct{ Message string }

func NewGitMergeConflictError(message string) *GitMergeConflictError {
	return &GitMergeConflictError{Message: message}
}
func (e *GitMergeConflictError) Error() string       { return e.Message }
func (e *GitMergeConflictError) HTTPStatusCode() int { return http.StatusConflict }
func (e *GitMergeConflictError) ErrorCode() string   { return string(CodeGitMergeConflict) }

type FileNotFoundError struct{ Message string }

func NewFileNotFoundError(message string) *FileNotFoundError {
	return &FileNotFoundError{Message: message}
}
func (e *FileNotFoundError) Error() string       { return e.Message }
func (e *FileNotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *FileNotFoundError) ErrorCode() string   { return string(CodeFileNotFound) }

type FileAccessDeniedError struct{ Message string }

func NewFileAccessDeniedError(message string) *FileAccessDeniedError {
	return &FileAccessDeniedError{Message: message}
}
func (e *FileAccessDeniedError) Error() string       { return e.Message }
func (e *FileAccessDeniedError) HTTPStatusCode() int { return http.StatusForbidden }
func (e *FileAccessDeniedError) ErrorCode() string   { return string(CodeFileAccessDenied) }

type LspServerNotInitializedError struct{ Message string }

func NewLspServerNotInitializedError(message string) *LspServerNotInitializedError {
	return &LspServerNotInitializedError{Message: message}
}
func (e *LspServerNotInitializedError) Error() string       { return e.Message }
func (e *LspServerNotInitializedError) HTTPStatusCode() int { return http.StatusBadRequest }
func (e *LspServerNotInitializedError) ErrorCode() string   { return string(CodeLspServerNotInitialized) }

type ProcessExecutionTimeoutError struct{ Message string }

func NewProcessExecutionTimeoutError(message string) *ProcessExecutionTimeoutError {
	return &ProcessExecutionTimeoutError{Message: message}
}
func (e *ProcessExecutionTimeoutError) Error() string       { return e.Message }
func (e *ProcessExecutionTimeoutError) HTTPStatusCode() int { return http.StatusRequestTimeout }
func (e *ProcessExecutionTimeoutError) ErrorCode() string   { return string(CodeProcessExecutionTimeout) }

type ProcessNotFoundError struct{ Message string }

func NewProcessNotFoundError(message string) *ProcessNotFoundError {
	return &ProcessNotFoundError{Message: message}
}
func (e *ProcessNotFoundError) Error() string       { return e.Message }
func (e *ProcessNotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *ProcessNotFoundError) ErrorCode() string   { return string(CodeProcessNotFound) }

type SessionEndedError struct{ Message string }

func NewSessionEndedError(message string) *SessionEndedError {
	return &SessionEndedError{Message: message}
}
func (e *SessionEndedError) Error() string       { return e.Message }
func (e *SessionEndedError) HTTPStatusCode() int { return http.StatusGone }
func (e *SessionEndedError) ErrorCode() string   { return string(CodeSessionEnded) }

type CommandAlreadyCompletedError struct{ Message string }

func NewCommandAlreadyCompletedError(message string) *CommandAlreadyCompletedError {
	return &CommandAlreadyCompletedError{Message: message}
}
func (e *CommandAlreadyCompletedError) Error() string       { return e.Message }
func (e *CommandAlreadyCompletedError) HTTPStatusCode() int { return http.StatusGone }
func (e *CommandAlreadyCompletedError) ErrorCode() string   { return string(CodeCommandAlreadyCompleted) }

type A11yUnavailableError struct{ Message string }

func NewA11yUnavailableError(message string) *A11yUnavailableError {
	return &A11yUnavailableError{Message: message}
}
func (e *A11yUnavailableError) Error() string       { return e.Message }
func (e *A11yUnavailableError) HTTPStatusCode() int { return http.StatusServiceUnavailable }
func (e *A11yUnavailableError) ErrorCode() string   { return string(CodeA11yUnavailable) }

type RecordingStillActiveError struct{ Message string }

func NewRecordingStillActiveError(message string) *RecordingStillActiveError {
	return &RecordingStillActiveError{Message: message}
}
func (e *RecordingStillActiveError) Error() string       { return e.Message }
func (e *RecordingStillActiveError) HTTPStatusCode() int { return http.StatusConflict }
func (e *RecordingStillActiveError) ErrorCode() string   { return string(CodeRecordingStillActive) }

type RecordingFfmpegNotFoundError struct{ Message string }

func NewRecordingFfmpegNotFoundError(message string) *RecordingFfmpegNotFoundError {
	return &RecordingFfmpegNotFoundError{Message: message}
}
func (e *RecordingFfmpegNotFoundError) Error() string       { return e.Message }
func (e *RecordingFfmpegNotFoundError) HTTPStatusCode() int { return http.StatusServiceUnavailable }
func (e *RecordingFfmpegNotFoundError) ErrorCode() string   { return string(CodeRecordingFfmpegNotFound) }

var (
	_ common_errors.HTTPError = (*GitAuthFailedError)(nil)
	_ common_errors.HTTPError = (*GitRepoNotFoundError)(nil)
	_ common_errors.HTTPError = (*GitBranchNotFoundError)(nil)
	_ common_errors.HTTPError = (*GitBranchExistsError)(nil)
	_ common_errors.HTTPError = (*GitPushRejectedError)(nil)
	_ common_errors.HTTPError = (*GitDirtyWorktreeError)(nil)
	_ common_errors.HTTPError = (*GitMergeConflictError)(nil)
	_ common_errors.HTTPError = (*FileNotFoundError)(nil)
	_ common_errors.HTTPError = (*FileAccessDeniedError)(nil)
	_ common_errors.HTTPError = (*LspServerNotInitializedError)(nil)
	_ common_errors.HTTPError = (*ProcessExecutionTimeoutError)(nil)
	_ common_errors.HTTPError = (*ProcessNotFoundError)(nil)
	_ common_errors.HTTPError = (*SessionEndedError)(nil)
	_ common_errors.HTTPError = (*CommandAlreadyCompletedError)(nil)
	_ common_errors.HTTPError = (*A11yUnavailableError)(nil)
	_ common_errors.HTTPError = (*RecordingStillActiveError)(nil)
	_ common_errors.HTTPError = (*RecordingFfmpegNotFoundError)(nil)
)
