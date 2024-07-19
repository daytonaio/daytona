// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"

	"strconv"

	"github.com/daytonaio/daytona/pkg/server"
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
	page, err := strconv.Atoi(ctx.Param("page"))
	if err != nil {
		page = 1
	}
	perPage, err := strconv.Atoi(ctx.Param("perPage"))
	if err != nil {
		perPage = 100
	}

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetRepositories(gitProviderId, namespaceId, page, perPage)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repositories for url: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
