// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/gitprovider/dto"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetGitContext 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git context
//	@Description	Get Git context
//	@Produce		json
//	@Param			repository	body		GetRepositoryContext	true	"Get repository context"
//	@Success		200			{object}	GitRepository
//	@Router			/gitprovider/context [post]
//
//	@id				GetGitContext
func GetGitContext(ctx *gin.Context) {
	var repositoryContext gitprovider.GetRepositoryContext
	if err := ctx.ShouldBindJSON(&repositoryContext); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to bind json: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	gitProvider, _, err := server.GitProviderService.GetGitProviderForUrl(repositoryContext.Url)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %s", err.Error()))
		return
	}

	repo, err := gitProvider.GetRepositoryContext(repositoryContext)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repository: %s", err.Error()))
		return
	}

	ctx.JSON(200, repo)
}

// GetUrlFromRepository 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get URL from Git repository
//	@Description	Get URL from Git repository
//	@Produce		json
//	@Param			repository	body		GitRepository	true	"Git repository"
//	@Success		200			{object}	RepositoryUrl
//	@Router			/gitprovider/context/url [post]
//
//	@id				GetUrlFromRepository
func GetUrlFromRepository(ctx *gin.Context) {
	var gitRepository gitprovider.GitRepository
	if err := ctx.ShouldBindJSON(&gitRepository); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to bind json: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	gitProvider, _, err := server.GitProviderService.GetGitProviderForUrl(gitRepository.Url)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for url: %s", err.Error()))
		return
	}

	url := gitProvider.GetUrlFromRepository(&gitRepository)

	response := dto.RepositoryUrl{
		URL: url,
	}

	ctx.JSON(200, response)
}
