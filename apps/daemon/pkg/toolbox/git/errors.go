// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func classifyGitError(err error) error {
	switch {
	case errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrInvalidAuthMethod):
		return common.NewGitAuthFailedError(err.Error())

	case errors.Is(err, transport.ErrRepositoryNotFound),
		errors.Is(err, go_git.ErrRepositoryNotExists):
		return common.NewGitRepoNotFoundError(err.Error())

	case errors.Is(err, go_git.ErrBranchNotFound):
		return common.NewGitBranchNotFoundError(err.Error())

	case errors.Is(err, go_git.ErrBranchExists):
		return common.NewGitBranchExistsError(err.Error())

	case errors.Is(err, go_git.ErrNonFastForwardUpdate):
		return common.NewGitPushRejectedError(err.Error())

	case errors.Is(err, go_git.ErrWorktreeNotClean),
		errors.Is(err, go_git.ErrUnstagedChanges):
		return common.NewGitDirtyWorktreeError(err.Error())

	case errors.Is(err, go_git.ErrFastForwardMergeNotPossible):
		return common.NewGitMergeConflictError(err.Error())

	case errors.Is(err, transport.ErrAuthorizationFailed):
		return common_errors.NewForbiddenError(err)

	case errors.Is(err, transport.ErrEmptyRemoteRepository),
		errors.Is(err, go_git.ErrRemoteNotFound),
		errors.Is(err, go_git.ErrTagNotFound),
		errors.Is(err, plumbing.ErrReferenceNotFound),
		errors.Is(err, plumbing.ErrObjectNotFound):
		return common_errors.NewNotFoundError(err)

	case errors.Is(err, go_git.ErrRepositoryAlreadyExists):
		return common_errors.NewConflictError(err)

	default:
		return common_errors.NewInternalServerError(err)
	}
}

func abortWithGitError(c *gin.Context, err error) {
	c.Error(classifyGitError(err))
}
