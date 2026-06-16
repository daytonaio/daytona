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

// RestoreFiles godoc
//
//	@Summary		Restore files
//	@Description	Restore working tree files or unstage changes
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitRestoreRequest	true	"Restore request"
//	@Success		200
//	@Router			/git/restore [post]
//
//	@id				RestoreFiles
func RestoreFiles(c *gin.Context) {
	var req GitRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	source := ""
	if req.Source != nil {
		source = *req.Source
	}

	if err := gitService.Restore(req.Files, req.Staged, req.Worktree, source); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
