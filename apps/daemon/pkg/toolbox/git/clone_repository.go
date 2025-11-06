// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/gin-gonic/gin"
	go_git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneRepository godoc
//
//	@Summary		Clone a Git repository
//	@Description	Clone a Git repository to the specified path
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitCloneRequest	true	"Clone repository request"
//	@Success		200
//	@Router			/git/clone [post]
//
//	@id				CloneRepository
func CloneRepository(c *gin.Context) {
	var req GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	branch := ""
	if req.Branch != nil {
		branch = *req.Branch
	}

	repo := gitprovider.GitRepository{
		Url:    req.URL,
		Branch: branch,
	}

	if req.CommitID != nil {
		repo.Target = gitprovider.CloneTargetCommit
		repo.Sha = *req.CommitID
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	var auth *go_git_http.BasicAuth

	// Set authentication if provided
	if req.Username != nil && req.Password != nil {
		auth = &go_git_http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	err := gitService.CloneRepository(&repo, auth)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusOK)
}
