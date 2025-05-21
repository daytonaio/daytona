// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

func ListBranches(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	gitService := git.Service{
		ProjectDir: path,
	}

	branchList, err := gitService.ListBranches()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, ListBranchResponse{
		Branches: branchList,
	})
}
