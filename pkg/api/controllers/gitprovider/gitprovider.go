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
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListGitProviders 			godoc
//
//	@Tags			gitProvider
//	@Summary		List Git providers
//	@Description	List Git providers
//	@Produce		json
//	@Success		200	{array}	models.GitProviderConfig
//	@Router			/gitprovider [get]
//
//	@id				ListGitProviders
func ListGitProviders(ctx *gin.Context) {
	var response []*models.GitProviderConfig

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.ListConfigs(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list git providers: %w", err))
		return
	}

	for _, provider := range response {
		provider.Token = ""
		provider.SigningKey = nil
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
//	@Success		200	{array}	models.GitProviderConfig
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

	gitProviders, err := server.GitProviderService.ListConfigsForUrl(ctx.Request.Context(), decodedUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %w", err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		for _, gitProvider := range gitProviders {
			gitProvider.Token = ""
		}
	}

	ctx.JSON(200, gitProviders)
}

// FindGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Find Git provider
//	@Description	Find Git provider
//	@Produce		plain
//	@Param			gitProviderId	path		string	true	"ID"
//	@Success		200				{object}	models.GitProviderConfig
//	@Router			/gitprovider/{gitProviderId} [get]
//
//	@id				FindGitProvider
func FindGitProvider(ctx *gin.Context) {
	id := ctx.Param("gitProviderId")

	server := server.GetInstance(nil)

	gitProvider, err := server.GitProviderService.FindConfig(ctx.Request.Context(), id)
	if err != nil {
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		gitProvider.Token = ""
	}

	ctx.JSON(200, gitProvider)
}

// FindGitProviderIdForUrl 			godoc
//
//	@Tags			gitProvider
//	@Summary		Find Git provider ID
//	@Description	Find Git provider ID
//	@Produce		plain
//	@Param			url	path		string	true	"Url"
//	@Success		200	{string}	providerId
//	@Router			/gitprovider/id-for-url/{url} [get]
//
//	@id				FindGitProviderIdForUrl
func FindGitProviderIdForUrl(ctx *gin.Context) {
	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, providerId, err := server.GitProviderService.GetGitProviderForUrl(ctx.Request.Context(), decodedUrl)
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

// SaveGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Save Git provider
//	@Description	Save Git provider
//	@Param			gitProviderConfig	body	SetGitProviderConfig	true	"Git provider"
//	@Produce		json
//	@Success		200
//	@Router			/gitprovider [put]
//
//	@id				SaveGitProvider
func SaveGitProvider(ctx *gin.Context) {
	var setConfigDto dto.SetGitProviderConfig

	err := ctx.BindJSON(&setConfigDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	gitProviderConfig := models.GitProviderConfig{
		Id:            setConfigDto.Id,
		ProviderId:    setConfigDto.ProviderId,
		Token:         setConfigDto.Token,
		BaseApiUrl:    setConfigDto.BaseApiUrl,
		SigningKey:    setConfigDto.SigningKey,
		SigningMethod: setConfigDto.SigningMethod,
	}

	if setConfigDto.Username != nil {
		gitProviderConfig.Username = *setConfigDto.Username
	}

	if setConfigDto.Alias != nil {
		gitProviderConfig.Alias = *setConfigDto.Alias
	}

	server := server.GetInstance(nil)

	err = server.GitProviderService.SaveGitProviderConfig(ctx.Request.Context(), &gitProviderConfig)
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

// DeleteGitProvider 			godoc
//
//	@Tags			gitProvider
//	@Summary		Delete Git provider
//	@Description	Delete Git provider
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Produce		json
//	@Success		200
//	@Router			/gitprovider/{gitProviderId} [delete]
//
//	@id				DeleteGitProvider
func DeleteGitProvider(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

	server := server.GetInstance(nil)

	err := server.GitProviderService.DeleteGitProviderConfig(ctx.Request.Context(), gitProviderId)
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
