// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetContainerRegistry godoc
//
//	@Tags			container-registry
//	@Summary		Set container registry credentials
//	@Description	Set container registry credentials
//	@Param			server				path	string				true	"Container Registry server name"
//	@Param			containerRegistry	body	ContainerRegistry	true	"Container Registry credentials to set"
//	@Success		201
//	@Router			/container-registry/{server} [put]
//
//	@id				SetContainerRegistry
func SetContainerRegistry(ctx *gin.Context) {
	crServer := ctx.Param("server")

	decodedServerURL, err := url.QueryUnescape(crServer)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode server URL: %w", err))
		return
	}

	var req containerregistry.ContainerRegistry
	err = ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	cr, err := server.ContainerRegistryService.Find(decodedServerURL)
	if err == nil {
		err = server.ContainerRegistryService.Delete(decodedServerURL)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove container registry: %w", err))
			return
		}

		cr.Server = req.Server
		cr.Username = req.Username
		cr.Password = req.Password
	} else {
		cr = &req
	}

	err = server.ContainerRegistryService.Save(cr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set container registry: %w", err))
		return
	}

	ctx.Status(201)
}
