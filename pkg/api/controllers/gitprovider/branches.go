// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/pkg/gitprovider"
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
	pageQuery := ctx.Query("page")
	perPageQuery := ctx.Query("per_page")

	namespaceId, err := url.QueryUnescape(namespaceArg)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to parse namespace: %s", err.Error()))
		return
	}

	repositoryId, err := url.QueryUnescape(repositoryArg)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to parse repository: %s", err.Error()))
		return
	}

	page := 1
	perPage := 100

	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for 'page' query param"))
			return
		}
	}

	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for 'per_page' query param"))
			return
		}
	}

	options := gitprovider.ListOptions{
		Page:    page,
		PerPage: perPage,
	}

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetRepoBranches(gitProviderId, namespaceId, repositoryId, options)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repo branches: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
