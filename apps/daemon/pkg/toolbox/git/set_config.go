// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// SetGitConfig godoc
//
//	@Summary		Set a Git config value
//	@Description	Set a Git config key/value at the given scope
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitSetConfigRequest	true	"Set config request"
//	@Success		200
//	@Router			/git/config [post]
//
//	@id				SetGitConfig
func SetGitConfig(c *gin.Context) {
	var req GitSetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	workDir := ""
	if req.Path != nil {
		workDir = *req.Path
	}
	gitService := git.Service{
		WorkDir: workDir,
	}

	scope := ""
	if req.Scope != nil {
		scope = *req.Scope
	}

	if scope == "local" && workDir == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("path is required when scope is local")))
		return
	}

	if err := gitService.SetConfigValue(req.Key, req.Value, scope); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
