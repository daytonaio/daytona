// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitChanges godoc
//
//	@Summary		Commit changes
//	@Description	Commit staged changes to the Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body		GitCommitRequest	true	"Commit request"
//	@Success		200		{object}	GitCommitResponse
//	@Router			/git/commit [post]
//
//	@id				CommitChanges
func CommitChanges(c *gin.Context) {
	var req GitCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	commitSha, err := gitService.Commit(req.Message, &go_git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author,
			Email: req.Email,
			When:  time.Now(),
		},
		AllowEmptyCommits: req.AllowEmpty,
	})

	if err != nil {
		abortWithGitError(c, err)
		return
	}

	c.JSON(http.StatusOK, GitCommitResponse{
		Hash: commitSha,
	})
}
