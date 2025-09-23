// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// CheckoutBranch godoc
//
//	@Summary		Checkout branch or commit
//	@Description	Switch to a different branch or commit in the Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitCheckoutRequest	true	"Checkout request"
//	@Success		200
//	@Router			/git/checkout [post]
//
//	@id				CheckoutBranch
func CheckoutBranch(c *gin.Context) {
	var req GitCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	if err := gitService.Checkout(req.Branch); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusOK)
}
