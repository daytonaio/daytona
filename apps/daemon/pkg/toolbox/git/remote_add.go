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

// AddRemote godoc
//
//	@Summary		Add a remote
//	@Description	Add (or overwrite) a remote in the Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitAddRemoteRequest	true	"Add remote request"
//	@Success		201
//	@Router			/git/remotes [post]
//
//	@id				AddRemote
func AddRemote(c *gin.Context) {
	var req GitAddRemoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	if err := gitService.AddRemote(req.Name, req.URL, req.Fetch, req.Overwrite); err != nil {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}
