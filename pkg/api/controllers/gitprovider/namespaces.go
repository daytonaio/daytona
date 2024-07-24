// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetNamespaces 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git namespaces
//	@Description	Get Git namespaces
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Param			page			query	int		false	"Page number"
//	@Param			per_page		query	int		false	"Number of items per page"
//	@Produce		json
//	@Success		200	{array}	GitNamespace
//	@Router			/gitprovider/{gitProviderId}/namespaces [get]
//
//	@id				GetNamespaces
func GetNamespaces(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")
	pageQuery := ctx.Query("page")
	perPageQuery := ctx.Query("per_page")

	var err error
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

	server := server.GetInstance(nil)

	options := gitprovider.ListOptions{
		Page:    page,
		PerPage: perPage,
	}

	response, err := server.GitProviderService.GetNamespaces(gitProviderId, options)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get namespaces: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
