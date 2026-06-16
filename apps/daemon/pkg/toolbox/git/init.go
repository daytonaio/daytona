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

// InitRepository godoc
//
//	@Summary		Initialize a Git repository
//	@Description	Initialize a new Git repository at the specified path
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitInitRequest	true	"Init repository request"
//	@Success		201
//	@Router			/git/init [post]
//
//	@id				InitRepository
func InitRepository(c *gin.Context) {
	var req GitInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	initialBranch := ""
	if req.InitialBranch != nil {
		initialBranch = *req.InitialBranch
	}

	if err := gitService.Init(req.Bare, initialBranch); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}
