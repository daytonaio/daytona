// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CloneRepository(c *gin.Context) {
	var req GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	options := &git.CloneOptions{
		URL:      req.URL,
		Progress: nil,
	}

	// Set authentication if provided
	if req.Username != nil && req.Password != nil {
		options.Auth = &http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	// Handle branch or commit specification
	if req.Branch != nil {
		options.ReferenceName = plumbing.NewBranchReferenceName(*req.Branch)
		options.SingleBranch = true
	}

	// Clone the repository
	repo, err := git.PlainClone(req.Path, false, options)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	// If a specific commit is requested, checkout that commit
	if req.CommitID != nil {
		worktree, err := repo.Worktree()
		if err != nil {
			c.AbortWithError(400, fmt.Errorf("failed to get worktree: %w", err))
			return
		}

		hash := plumbing.NewHash(*req.CommitID)
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: hash,
		})
		if err != nil {
			c.AbortWithError(400, fmt.Errorf("failed to checkout commit: %w", err))
			return
		}
	}

	c.Status(200)
}
