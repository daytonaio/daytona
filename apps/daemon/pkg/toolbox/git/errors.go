// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func classifyGitError(err error) error {
	if errors.Is(err, transport.ErrAuthenticationRequired) ||
		errors.Is(err, transport.ErrInvalidAuthMethod) {
		return common_errors.NewUnauthorizedError(err)
	}

	if errors.Is(err, transport.ErrAuthorizationFailed) {
		return common_errors.NewForbiddenError(err)
	}

	if errors.Is(err, transport.ErrRepositoryNotFound) ||
		errors.Is(err, transport.ErrEmptyRemoteRepository) ||
		errors.Is(err, go_git.ErrRepositoryNotExists) ||
		errors.Is(err, go_git.ErrBranchNotFound) ||
		errors.Is(err, go_git.ErrRemoteNotFound) ||
		errors.Is(err, go_git.ErrTagNotFound) ||
		errors.Is(err, plumbing.ErrReferenceNotFound) ||
		errors.Is(err, plumbing.ErrObjectNotFound) {
		return common_errors.NewNotFoundError(err)
	}

	if errors.Is(err, go_git.ErrNonFastForwardUpdate) ||
		errors.Is(err, go_git.ErrWorktreeNotClean) ||
		errors.Is(err, go_git.ErrUnstagedChanges) ||
		errors.Is(err, go_git.ErrRepositoryAlreadyExists) ||
		errors.Is(err, go_git.ErrBranchExists) ||
		errors.Is(err, go_git.ErrFastForwardMergeNotPossible) {
		return common_errors.NewConflictError(err)
	}

	return common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR")
}

func abortWithGitError(c *gin.Context, err error) {
	_ = c.Error(classifyGitError(err))
	c.Abort()
}
