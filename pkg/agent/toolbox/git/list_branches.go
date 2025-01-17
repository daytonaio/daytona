// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/git"
	"github.com/gin-gonic/gin"
)

func ListBranches(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	gitService := git.Service{
		WorkspaceDir: path,
	}

	branchList, err := gitService.ListBranches()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, ListBranchResponse{
		Branches: branchList,
	})
}
