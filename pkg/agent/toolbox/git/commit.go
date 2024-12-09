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
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	repo, err := git.PlainOpen(req.Path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = worktree.Add(".")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"hash": commit.String()})
}
