// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/gin-gonic/gin"
)

// GetRepositories 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git repositories
//	@Description	Get Git repositories
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Param			namespaceId		path	string	true	"Namespace"
//	@Produce		json
//	@Success		200	{array}	GitRepository
//	@Router			/gitprovider/{gitProviderId}/{namespaceId}/repositories [get]
//
//	@id				GetRepositories
func GetRepositories(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")
	namespaceId := ctx.Param("namespaceId")

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	gitProvider := gitprovider.GetGitProvider(gitProviderId, c.GitProviders)
	if gitProvider == nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("git provider not found"))
		return
	}

	response, err := gitProvider.GetRepositories(namespaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repositories: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
