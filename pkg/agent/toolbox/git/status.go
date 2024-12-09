// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
)

func GetStatus(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(400, gin.H{"error": "path is required"})
		return
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	status, err := worktree.Status()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, status)
}
