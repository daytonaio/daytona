// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func PushChanges(c *gin.Context) {
	var req GitPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	repo, err := git.PlainOpen(req.Path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	options := &git.PushOptions{}
	if req.Username != "" && req.Password != "" {
		options.Auth = &http.BasicAuth{
			Username: req.Username,
			Password: req.Password,
		}
	}

	err = repo.Push(options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Changes pushed successfully"})
}
