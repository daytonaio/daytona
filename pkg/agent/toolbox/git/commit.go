// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"time"

	"github.com/daytonaio/daytona/pkg/git"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func CommitChanges(c *gin.Context) {
	var req GitCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	gitService := git.Service{
		WorkspaceDir: req.Path,
	}

	commitSha, err := gitService.Commit(req.Message, &go_git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author,
			Email: req.Email,
			When:  time.Now(),
		},
	})

	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, GitCommitResponse{
		Hash: commitSha,
	})
}
