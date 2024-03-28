// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetGitContext 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git context
//	@Description	Get Git context
//	@Produce		json
//	@Param			gitUrl	path		string	true	"Git URL"
//	@Success		200		{object}	GitRepository
//	@Router			/gitprovider/context/{gitUrl} [get]
//
//	@id				GetGitContext
func GetGitContext(ctx *gin.Context) {
	// TODO: needs real implementing
	gitUrl := ctx.Param("gitUrl")

	decodedURLParam, err := url.QueryUnescape(gitUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	repo := &types.GitRepository{}
	repo.Url = decodedURLParam

	ctx.JSON(200, repo)
}
