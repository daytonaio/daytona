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

// ListRemotes godoc
//
//	@Summary		List remotes
//	@Description	List the remotes configured in the Git repository
//	@Tags			git
//	@Produce		json
//	@Param			path	query		string	true	"Repository path"
//	@Success		200		{object}	ListRemotesResponse
//	@Router			/git/remotes [get]
//
//	@id				ListRemotes
func ListRemotes(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		_ = c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	gitService := git.Service{
		WorkDir: path,
	}

	remotes, err := gitService.ListRemotes()
	if err != nil {
		abortWithGitError(c, err)
		return
	}

	response := ListRemotesResponse{
		Remotes: make([]GitRemote, 0, len(remotes)),
	}
	for _, remote := range remotes {
		response.Remotes = append(response.Remotes, GitRemote{
			Name: remote.Name,
			URL:  remote.URL,
		})
	}

	c.JSON(http.StatusOK, response)
}
