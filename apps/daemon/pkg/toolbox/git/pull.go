// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	go_git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// PullChanges godoc
//
//	@Summary		Pull changes from remote
//	@Description	Pull changes from the remote Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitRepoRequest	true	"Pull request"
//	@Success		200
//	@Router			/git/pull [post]
//
//	@id				PullChanges
func PullChanges(c *gin.Context) {
	var req GitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	var auth *go_git_http.BasicAuth
	if req.Username != nil && req.Password != nil {
		auth = &go_git_http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	err := gitService.Pull(auth)
	if err != nil && err != go_git.NoErrAlreadyUpToDate {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusOK)
}
