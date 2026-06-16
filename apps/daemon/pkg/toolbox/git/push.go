// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
	go_git "github.com/go-git/go-git/v5"
	go_git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// PushChanges godoc
//
//	@Summary		Push changes to remote
//	@Description	Push local changes to the remote Git repository
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body	GitPushRequest	true	"Push request"
//	@Success		200
//	@Router			/git/push [post]
//
//	@id				PushChanges
func PushChanges(c *gin.Context) {
	var req GitPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	var auth *go_git_http.BasicAuth
	if req.Username != nil && req.Password != nil {
		auth = &go_git_http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	remote := ""
	if req.Remote != nil {
		remote = *req.Remote
	}
	branch := ""
	if req.Branch != nil {
		branch = *req.Branch
	}
	setUpstream := req.SetUpstream != nil && *req.SetUpstream

	err := gitService.Push(auth, remote, branch, setUpstream)
	if err != nil && err != go_git.NoErrAlreadyUpToDate {
		abortWithGitError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
