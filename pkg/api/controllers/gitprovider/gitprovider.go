// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"net/url"

	"github.com/daytonaio/daytona/pkg/api/controllers/gitprovider/dto"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListGitProviders 			godoc
//
//	@Tags			gitProvider
//	@Summary		List Git providers
//	@Description	List Git providers
//	@Produce		json
//	@Success		200	{array}	gitprovider.GitProviderConfig
//	@Router			/gitprovider [get]
//
//	@id				ListGitProviders
func ListGitProviders(ctx *gin.Context) {
	var response []*gitprovider.GitProviderConfig

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.ListConfigs()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list git providers: %s", err.Error()))
		return
	}

	for _, provider := range response {
		provider.Token = ""
	}

	ctx.JSON(200, response)
}

// GetGitProviderForUrl 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git provider
//	@Description	Get Git provider
//	@Produce		json
//	@Param			url	path		string	true	"Url"
//	@Success		200	{object}	gitprovider.GitProviderConfig
//	@Router			/gitprovider/for-url/{url} [get]
//
//	@id				GetGitProviderForUrl
func GetGitProviderForUrl(ctx *gin.Context) {
	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	gitProvider, err := server.GitProviderService.GetConfigForUrl(decodedUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %s", err.Error()))
		return
	}

	ctx.JSON(200, gitProvider)
}

// GetGitProviderIdForUrl 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git provider ID
//	@Description	Get Git provider ID
//	@Produce		plain
//	@Param			url	path		string	true	"Url"
//	@Success		200	{string}	providerId
//	@Router			/gitprovider/id-for-url/{url} [get]
//
//	@id				GetGitProviderIdForUrl
func GetGitProviderIdForUrl(ctx *gin.Context) {
	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	_, providerId, err := server.GitProviderService.GetGitProviderForUrl(decodedUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %s", err.Error()))
		return
	}

	ctx.String(200, providerId)
}

// SetGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Set Git provider
//	@Description	Set Git provider
//	@Param			gitProviderConfig	body	SetGitProviderConfig	true	"Git provider"
//	@Produce		json
//	@Success		200
//	@Router			/gitprovider [put]
//
//	@id				SetGitProvider
func SetGitProvider(ctx *gin.Context) {
	var setConfigDto dto.SetGitProviderConfig

	err := ctx.BindJSON(&setConfigDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	gitProviderConfig := gitprovider.GitProviderConfig{
		Id:             setConfigDto.Id,
		Token:          setConfigDto.Token,
		BaseApiUrl:     setConfigDto.BaseApiUrl,
		TokenIdentity:  setConfigDto.TokenIdentity,
		TokenScope:     setConfigDto.TokenScope,
		TokenScopeType: setConfigDto.TokenScopeType,
	}

	if setConfigDto.Username != nil {
		gitProviderConfig.Username = *setConfigDto.Username
	}

	server := server.GetInstance(nil)

	err = server.GitProviderService.SetGitProviderConfig(&gitProviderConfig)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to set git provider: %s", err.Error()))
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

	server := server.GetInstance(nil)

	err := server.GitProviderService.RemoveGitProvider(gitProviderId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove git provider: %s", err.Error()))
		return
	}

	ctx.JSON(200, nil)
}
