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

// GetRepoPRs 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git repository PRs
//	@Description	Get Git repository PRs
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Param			namespaceId		path	string	true	"Namespace"
//	@Param			repositoryId	path	string	true	"Repository"
//	@Produce		json
//	@Success		200	{array}	GitPullRequest
//	@Router			/gitprovider/{gitProviderId}/{namespaceId}/{repositoryId}/pull-requests [get]
//
//	@id				GetRepoPRs
func GetRepoPRs(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")
	namespaceId := ctx.Param("namespaceId")
	repositoryId := ctx.Param("repositoryId")

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

	response, err := gitProvider.GetRepoPRs(repositoryId, namespaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get pull requests: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
