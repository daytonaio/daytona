// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"net/url"

	"github.com/daytonaio/daytona/pkg/api/controllers"
	"github.com/daytonaio/daytona/pkg/api/controllers/gitprovider/dto"
	"github.com/daytonaio/daytona/pkg/apikey"
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
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list git providers: %w", err))
		return
	}

	for _, provider := range response {
		provider.Token = ""
	}

	ctx.JSON(200, response)
}

// ListGitProvidersForUrl 			godoc
//
//	@Tags			gitProvider
//	@Summary		List Git providers for url
//	@Description	List Git providers for url
//	@Produce		json
//	@Param			url	path	string	true	"Url"
//	@Success		200	{array}	gitprovider.GitProviderConfig
//	@Router			/gitprovider/for-url/{url} [get]
//
//	@id				ListGitProvidersForUrl
func ListGitProvidersForUrl(ctx *gin.Context) {
	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %w", err))
		return
	}

	server := server.GetInstance(nil)

	gitProviders, err := server.GitProviderService.ListConfigsForUrl(decodedUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %w", err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == apikey.ApiKeyTypeClient {
		for _, gitProvider := range gitProviders {
			gitProvider.Token = ""
		}
	}

	ctx.JSON(200, gitProviders)
}

// GetGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git provider
//	@Description	Get Git provider
//	@Produce		plain
//	@Param			gitProviderId	path		string	true	"ID"
//	@Success		200				{object}	gitprovider.GitProviderConfig
//	@Router			/gitprovider/{gitProviderId} [get]
//
//	@id				GetGitProvider
func GetGitProvider(ctx *gin.Context) {
	id := ctx.Param("gitProviderId")

	server := server.GetInstance(nil)

	gitProvider, err := server.GitProviderService.GetConfig(id)
	if err != nil {
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
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
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, providerId, err := server.GitProviderService.GetGitProviderForUrl(decodedUrl)
	if err != nil {
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
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
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitProviderConfig := gitprovider.GitProviderConfig{
		Id:         setConfigDto.Id,
		ProviderId: setConfigDto.ProviderId,
		Token:      setConfigDto.Token,
		BaseApiUrl: setConfigDto.BaseApiUrl,
	}

	if setConfigDto.Username != nil {
		gitProviderConfig.Username = *setConfigDto.Username
	}

	if setConfigDto.Alias != nil {
		gitProviderConfig.Alias = *setConfigDto.Alias
	}

	server := server.GetInstance(nil)

	err = server.GitProviderService.SetGitProviderConfig(&gitProviderConfig)
	if err != nil {
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
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
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
		return
	}

	ctx.JSON(200, nil)
}

// extract pagination related query params
func getListOptions(ctx *gin.Context) (gitprovider.ListOptions, error) {
	pageQuery := ctx.Query("page")
	perPageQuery := ctx.Query("per_page")

	page := 1
	perPage := 100
	var err error

	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			return gitprovider.ListOptions{}, errors.New("invalid value for 'page' query param")
		}
	}

	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			return gitprovider.ListOptions{}, errors.New("invalid value for 'per_page' query param")
		}
	}

	return gitprovider.ListOptions{
		Page:    page,
		PerPage: perPage,
	}, nil
}
