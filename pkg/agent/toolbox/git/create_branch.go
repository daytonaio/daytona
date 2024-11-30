// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func CreateBranch(c *gin.Context) {
	var req GitBranchRequest
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

	err = worktree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName(req.Name),
	})

	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(201)
}
