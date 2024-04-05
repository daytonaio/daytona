// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"net/url"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/gin-gonic/gin"
)

// GetGitProviderForUrl 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git provider
//	@Description	Get Git provider
//	@Produce		json
//	@Param			url	path		string	true	"Url"
//	@Success		200	{object}	types.GitProvider
//	@Router			/gitprovider/for-url/{url} [get]
//
//	@id				GetGitProviderForUrl
func GetGitProviderForUrl(ctx *gin.Context) {
	var response gitprovider.GitProvider

	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	response, err = gitproviders.GetGitProviderForUrl(decodedUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}

// SetGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Set Git provider
//	@Description	Set Git provider
//	@Param			gitProviderData	body	types.GitProvider	true	"Git provider"
//	@Produce		json
//	@Success		200
//	@Router			/gitprovider [put]
//
//	@id				SetGitProvider
func SetGitProvider(ctx *gin.Context) {
	var gitProviderData gitprovider.GitProvider

	err := ctx.BindJSON(&gitProviderData)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	err = gitproviders.SetGitProvider(gitProviderData)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set git provider: %s", err.Error()))
		return
	}

	ctx.JSON(200, nil)
}

// RemoveGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Remove Git provider
//	@Description	Remove Git provider
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Produce		json
//	@Success		200
//	@Router			/gitprovider/{gitProviderId} [delete]
//
//	@id				RemoveGitProvider
func RemoveGitProvider(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

	err := gitproviders.RemoveGitProvider(gitProviderId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove git provider: %s", err.Error()))
		return
	}

	ctx.JSON(200, nil)
}
