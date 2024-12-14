// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func PushChanges(c *gin.Context) {
	var req GitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	var auth *http.BasicAuth
	if req.Username != nil && req.Password != nil {
		auth = &http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	gitService := git.Service{
		ProjectDir: req.Path,
	}

	err := gitService.Push(auth)
	if err != nil && err != go_git.NoErrAlreadyUpToDate {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}
