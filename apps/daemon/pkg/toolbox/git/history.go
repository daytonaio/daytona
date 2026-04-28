// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/gin-gonic/gin"
)

// GetCommitHistory godoc
//
//	@Summary		Get commit history
//	@Description	Get the commit history of the Git repository
//	@Tags			git
//	@Produce		json
//	@Param			path	query	string	true	"Repository path"
//	@Success		200		{array}	git.GitCommitInfo
//	@Router			/git/history [get]
//
//	@id				GetCommitHistory
func GetCommitHistory(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	gitService := git.Service{
		WorkDir: path,
	}

	log, err := gitService.Log()
	if err != nil {
		abortWithGitError(c, err)
		return
	}

	c.JSON(http.StatusOK, log)
}
