// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// CreateBranch godoc
//
//	@Summary		Create a new branch
//	@Description	Create a new branch in the Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitBranchRequest	true	"Create branch request"
//	@Success		201
//	@Router			/git/branches [post]
//
//	@id				CreateBranch
func CreateBranch(c *gin.Context) {
	var req GitBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	if err := gitService.CreateBranch(req.Name); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusCreated)
}
