// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func ListBranches(c *gin.Context) {
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

	branches, err := repo.Branches()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var branchList []string
	branches.ForEach(func(ref *plumbing.Reference) error {
		branchList = append(branchList, ref.Name().Short())
		return nil
	})

	c.JSON(200, branchList)
}
