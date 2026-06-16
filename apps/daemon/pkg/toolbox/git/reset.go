// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// ResetChanges godoc
//
//	@Summary		Reset repository
//	@Description	Reset the current HEAD to the specified state
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitResetRequest	true	"Reset request"
//	@Success		200
//	@Router			/git/reset [post]
//
//	@id				ResetChanges
func ResetChanges(c *gin.Context) {
	var req GitResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	mode := ""
	if req.Mode != nil {
		mode = *req.Mode
	}
	target := ""
	if req.Target != nil {
		target = *req.Target
	}

	if err := gitService.Reset(mode, target, req.Files); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
