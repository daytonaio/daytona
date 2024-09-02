// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetRepoBranches 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git repository branches
//	@Description	Get Git repository branches
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Param			namespaceId		path	string	true	"Namespace"
//	@Param			repositoryId	path	string	true	"Repository"
//	@Produce		json
//	@Success		200	{array}	GitBranch
//	@Router			/gitprovider/{gitProviderId}/{namespaceId}/{repositoryId}/branches [get]
//
//	@id				GetRepoBranches
func GetRepoBranches(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")
	namespaceArg := ctx.Param("namespaceId")
	repositoryArg := ctx.Param("repositoryId")

	namespaceId, err := url.QueryUnescape(namespaceArg)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to parse namespace: %w", err))
		return
	}

	repositoryId, err := url.QueryUnescape(repositoryArg)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to parse repository: %w", err))
		return
	}

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetRepoBranches(gitProviderId, namespaceId, repositoryId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repo branches: %w", err))
		return
	}

	ctx.JSON(200, response)
}
