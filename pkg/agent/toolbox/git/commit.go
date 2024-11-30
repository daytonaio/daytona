// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func CommitChanges(c *gin.Context) {
	var req GitCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	repo, err := git.PlainOpen(req.Path)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	commit, err := worktree.Commit(req.Message, &git.CommitOptions{
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
		Hash: commit.String(),
	})
}
