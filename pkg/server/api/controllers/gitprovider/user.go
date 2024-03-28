// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetGitUser 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git context
//	@Description	Get Git context
//	@Produce		json
//	@Param			gitProviderId	path		string	true	"Git Provider Id"
//	@Success		200				{object}	GitUser
//	@Router			/gitprovider/{gitProviderId}/user [get]
//
//	@id				GetGitUser
func GetGitUser(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	gitProviderInstance := gitprovider.GetGitProvider(gitProviderId, c.GitProviders)
	if gitProviderInstance == nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("git provider not found"))
		return
	}

	user, err := gitProviderInstance.GetUser()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get user: %s", err.Error()))
		return
	}

	ctx.JSON(200, user)
}

// GetGitUsernameFromToken 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get username from token
//	@Description	Get username from token
//	@Produce		json
//	@Param			gitProviderData	body		types.GitProvider	true	"Git provider"
//	@Success		200				{string}	username
//	@Router			/gitprovider/username-from-token [get]
//
//	@id				GetGitUsernameFromToken
func GetGitUsernameFromToken(ctx *gin.Context) {
	var gitProviderData types.GitProvider

	err := ctx.BindJSON(&gitProviderData)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to bind json: %s", err.Error()))
		return
	}

	username, err := gitprovider.GetUsernameFromToken(gitProviderData.Id, gitProviderData.Token, gitProviderData.BaseApiUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get user: %s", err.Error()))
		return
	}

	ctx.String(200, username)
}
