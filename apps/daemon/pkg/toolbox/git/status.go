// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// GetStatus godoc
//
//	@Summary		Get Git status
//	@Description	Get the Git status of the repository at the specified path
//	@Tags			git
//	@Produce		json
//	@Param			path	query		string	true	"Repository path"
//	@Success		200		{object}	git.GitStatus
//	@Router			/git/status [get]
//
//	@id				GetStatus
func GetStatus(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	gitService := git.Service{
		WorkDir: path,
	}

	status, err := gitService.GetGitStatus()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, status)
}
