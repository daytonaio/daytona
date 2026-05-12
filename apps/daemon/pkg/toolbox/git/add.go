// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// AddFiles godoc
//
//	@Summary		Add files to Git staging
//	@Description	Add files to the Git staging area
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitAddRequest	true	"Add files request"
//	@Success		200
//	@Router			/git/add [post]
//
//	@id				AddFiles
func AddFiles(c *gin.Context) {
	var req GitAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	if err := gitService.Add(req.Files); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
