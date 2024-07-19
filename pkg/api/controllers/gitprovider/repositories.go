// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
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
//	@Param			page			query	integer	false	"Page"
//	@Param			perPage			query	integer	false	"Per page"
//	@Produce		json
//	@Success		200	{array}	GitRepository
//	@Router			/gitprovider/{gitProviderId}/{namespaceId}/repositories [get]
//
//	@id				GetRepositories
func GetRepositories(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")
	namespaceId := ctx.Param("namespaceId")
	pageQuery := ctx.Query("page")
	perPageQuery := ctx.Query("per_page")

	var err error
	page := 1
	perPage := 100

	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for page flag"))
			return
		}
	}

	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for per_page flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetRepositories(gitProviderId, namespaceId, page, perPage)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get repositories for url: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
