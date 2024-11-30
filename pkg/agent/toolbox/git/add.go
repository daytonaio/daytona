// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
)

func AddFiles(c *gin.Context) {
	var req GitAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	repo, err := git.PlainOpen(req.Path)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if len(req.Files) == 1 && req.Files[0] == "." {
		_, err = w.Add(".")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
	} else {
		for _, file := range req.Files {
			_, err = w.Add(file)
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		}
	}

	c.Status(200)
}
