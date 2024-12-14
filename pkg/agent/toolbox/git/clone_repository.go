// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CloneRepository(c *gin.Context) {
	var req GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	branch := "main"
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
		ProjectDir: req.Path,
	}

	var auth *http.BasicAuth

	// Set authentication if provided
	if req.Username != nil && req.Password != nil {
		auth = &http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	err := gitService.CloneRepository(&repo, auth)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}
