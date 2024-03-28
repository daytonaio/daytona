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

// GetNamespaces 			godoc
//
//	@Tags			gitProvider
//	@Summary		Get Git namespaces
//	@Description	Get Git namespaces
//	@Param			gitProviderId	path	string	true	"Git provider"
//	@Produce		json
//	@Success		200	{array}	GitNamespace
//	@Router			/gitprovider/{gitProviderId}/namespaces [get]
//
//	@id				GetNamespaces
func GetNamespaces(ctx *gin.Context) {
	gitProviderId := ctx.Param("gitProviderId")

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

	response, err := gitProvider.GetNamespaces()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get namespaces: %s", err.Error()))
		return
	}

	ctx.JSON(200, response)
}
