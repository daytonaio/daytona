// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"

	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers"

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
	options, err := getListOptions(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	server := server.GetInstance(nil)

	response, err := server.GitProviderService.GetNamespaces(gitProviderId, options)
	if err != nil {
		statusCode, message, codeErr := controllers.GetHTTPStatusCodeAndMessageFromError(err)
		if codeErr != nil {
			ctx.AbortWithError(statusCode, codeErr)
		}
		ctx.AbortWithError(statusCode, errors.New(message))
		return
	}

	ctx.JSON(200, response)
}
