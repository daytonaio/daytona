// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitCloneRequest struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	Username string `json:"username"`
	Password string `json:"password"`
	Branch   string `json:"branch"`
	CommitID string `json:"commit_id"`
} // @name GitCloneRequest

func CloneRepository(c *gin.Context) {
	var req GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	options := &git.CloneOptions{
		URL:      req.URL,
		Progress: nil,
	}

	// Set authentication if provided
	if req.Username != "" && req.Password != "" {
		options.Auth = &http.BasicAuth{
			Username: req.Username,
			Password: req.Password,
		}
	}

	// Handle branch or commit specification
	if req.Branch != "" {
		options.ReferenceName = plumbing.NewBranchReferenceName(req.Branch)
		options.SingleBranch = true
	}

	// Clone the repository
	repo, err := git.PlainClone(req.Path, false, options)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// If a specific commit is requested, checkout that commit
	if req.CommitID != "" {
		worktree, err := repo.Worktree()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get worktree: " + err.Error()})
			return
		}

		hash := plumbing.NewHash(req.CommitID)
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: hash,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to checkout commit: " + err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Repository cloned successfully"})
}
