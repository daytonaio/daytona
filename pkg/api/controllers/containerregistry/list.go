// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetContainerRegistry godoc
//
//	@Tags			container-registry
//	@Summary		Get container registry credentials
//	@Description	Get container registry credentials
//	@Produce		json
//	@Param			server	path		string	true	"Container Registry server name"
//	@Success		200		{object}	ContainerRegistry
//	@Router			/container-registry/{server} [get]
//
//	@id				GetContainerRegistry
func GetContainerRegistry(ctx *gin.Context) {
	crServer := ctx.Param("server")

	decodedServerURL, err := url.QueryUnescape(crServer)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode server URL: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	cr, err := server.ContainerRegistryService.Find(decodedServerURL)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get container registry: %s", err.Error()))
		return
	}

	cr.Password = ""

	ctx.JSON(200, cr)
}

// ListContainerRegistries godoc
//
//	@Tags			container-registry
//	@Summary		List container registries
//	@Description	List container registries
//	@Produce		json
//	@Success		200	{array}	ContainerRegistry
//	@Router			/container-registry [get]
//
//	@id				ListContainerRegistries
func ListContainerRegistries(ctx *gin.Context) {
	server := server.GetInstance(nil)

	crs, err := server.ContainerRegistryService.List()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list container registries: %s", err.Error()))
		return
	}

	for _, cr := range crs {
		cr.Password = ""
	}

	ctx.JSON(200, crs)
}
