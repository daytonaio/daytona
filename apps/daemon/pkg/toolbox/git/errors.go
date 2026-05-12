// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	"github.com/daytonaio/daemon/pkg/common"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// GitErrorResponse is the error response shape for all git endpoints.
// Used only for swaggo documentation — the actual response is written by the middleware.
type GitErrorResponse = common.ErrorResponse // @name GitErrorResponse

func classifyGitError(err error) error {
	var code common.DaemonErrorCode
	var statusCode int

	switch {
	case errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrInvalidAuthMethod):
		code, statusCode = common.CodeGitAuthFailed, http.StatusUnauthorized

	case errors.Is(err, transport.ErrAuthorizationFailed):
		code, statusCode = common.CodeGitAuthForbidden, http.StatusForbidden

	case errors.Is(err, transport.ErrRepositoryNotFound),
		errors.Is(err, go_git.ErrRepositoryNotExists):
		code, statusCode = common.CodeGitRepoNotFound, http.StatusNotFound

	case errors.Is(err, transport.ErrEmptyRemoteRepository):
		code, statusCode = common.CodeGitEmptyRepo, http.StatusNotFound

	case errors.Is(err, go_git.ErrBranchNotFound):
		code, statusCode = common.CodeGitBranchNotFound, http.StatusNotFound

	case errors.Is(err, go_git.ErrRemoteNotFound),
		errors.Is(err, go_git.ErrTagNotFound),
		errors.Is(err, plumbing.ErrReferenceNotFound),
		errors.Is(err, plumbing.ErrObjectNotFound):
		code, statusCode = common.CodeGitRefNotFound, http.StatusNotFound

	case errors.Is(err, go_git.ErrNonFastForwardUpdate):
		code, statusCode = common.CodeGitPushRejected, http.StatusConflict

	case errors.Is(err, go_git.ErrWorktreeNotClean),
		errors.Is(err, go_git.ErrUnstagedChanges):
		code, statusCode = common.CodeGitDirtyWorktree, http.StatusConflict

	case errors.Is(err, go_git.ErrRepositoryAlreadyExists),
		errors.Is(err, go_git.ErrBranchExists):
		code, statusCode = common.CodeGitBranchExists, http.StatusConflict

	case errors.Is(err, go_git.ErrFastForwardMergeNotPossible):
		code, statusCode = common.CodeGitMergeConflict, http.StatusConflict

	default:
		code, statusCode = common.CodeInternalServerError, http.StatusInternalServerError
	}

	return common_errors.NewCustomError(statusCode, err.Error(), string(code))
}

func abortWithGitError(c *gin.Context, err error) {
	_ = c.Error(classifyGitError(err))
	c.Abort()
}
