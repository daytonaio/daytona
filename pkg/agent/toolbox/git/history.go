// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func GetCommitHistory(c *gin.Context) {
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

	ref, err := repo.Head()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	commits, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var history []GitCommitInfo
	commits.ForEach(func(commit *object.Commit) error {
		history = append(history, GitCommitInfo{
			Hash:      commit.Hash.String(),
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Message:   commit.Message,
			Timestamp: commit.Author.When,
		})
		return nil
	})

	c.JSON(200, history)
}
