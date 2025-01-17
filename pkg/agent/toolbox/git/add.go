// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/gin-gonic/gin"
)

func AddFiles(c *gin.Context) {
	var req GitAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	gitService := git.Service{
		WorkspaceDir: req.Path,
	}

	if err := gitService.Add(req.Files); err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}
