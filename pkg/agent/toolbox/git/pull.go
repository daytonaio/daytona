// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func PullChanges(c *gin.Context) {
	var req GitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	repo, err := git.PlainOpen(req.Path)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	options := &git.PullOptions{
		RemoteName: "origin",
	}

	if req.Username != nil && req.Password != nil {
		options.Auth = &http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	err = w.Pull(options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}
