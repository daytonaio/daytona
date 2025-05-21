// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

func CreateBranch(c *gin.Context) {
	var req GitBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitService := git.Service{
		ProjectDir: req.Path,
	}

	if err := gitService.CreateBranch(req.Name); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusCreated)
}
