// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func ListBranches(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	branches, err := repo.Branches()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	var branchList []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchList = append(branchList, ref.Name().Short())
		return nil
	})
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, ListBranchResponse{
		Branches: branchList,
	})
}
