// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// DeleteBranch godoc
//
//	@Summary		Delete a branch
//	@Description	Delete a branch from the Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitDeleteBranchRequest	true	"Delete branch request"
//	@Success		204
//	@Router			/git/branches [delete]
//
//	@id				DeleteBranch
func DeleteBranch(c *gin.Context) {
	var req GitDeleteBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	if err := gitService.DeleteBranch(req.Name); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}
