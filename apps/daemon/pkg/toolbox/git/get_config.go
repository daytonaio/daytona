// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// GetGitConfig godoc
//
//	@Summary		Get a Git config value
//	@Description	Get a Git config value at the given scope (null when unset)
//	@Tags			git
//	@Produce		json
//	@Param			key		query		string	true	"Config key (e.g. user.name)"
//	@Param			path	query		string	false	"Repository path (required for local scope)"
//	@Param			scope	query		string	false	"Config scope: global (default), local or system"
//	@Success		200		{object}	GitConfigResponse
//	@Router			/git/config [get]
//
//	@id				GetGitConfig
func GetGitConfig(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("key is required")))
		return
	}

	if c.Query("scope") == "local" && c.Query("path") == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("path is required when scope is local")))
		return
	}

	gitService := git.Service{
		WorkDir: c.Query("path"),
	}

	value, err := gitService.GetConfigValue(key, c.Query("scope"))
	if err != nil {
		abortWithGitError(c, err)
		return
	}

	c.JSON(http.StatusOK, GitConfigResponse{
		Value: value,
	})
}
