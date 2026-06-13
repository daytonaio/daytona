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

// ConfigureUser godoc
//
//	@Summary		Configure Git user
//	@Description	Configure the Git user name and email at the given scope
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitConfigureUserRequest	true	"Configure user request"
//	@Success		200
//	@Router			/git/config/user [post]
//
//	@id				ConfigureUser
func ConfigureUser(c *gin.Context) {
	var req GitConfigureUserRequest
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

	if err := gitService.ConfigureUser(req.Name, req.Email, scope); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
