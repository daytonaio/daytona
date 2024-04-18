// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RemoveContainerRegistry godoc
//
//	@Tags			container-registry
//	@Summary		Remove a container registry credentials
//	@Description	Remove a container registry credentials
//	@Param			server path	string	true	"Container Registry server name"
//	@Param			username	path		string	true	"Container Registry username"
//	@Success		204
//	@Router			/container-registry/{server}/{username} [delete]
//
//	@id				RemoveContainerRegistry
func RemoveContainerRegistry(ctx *gin.Context) {
	crServer := ctx.Param("server")
	crUsername := ctx.Param("username")

	server := server.GetInstance(nil)

	err := server.ContainerRegistryService.Delete(crServer, crUsername)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove container registry: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
