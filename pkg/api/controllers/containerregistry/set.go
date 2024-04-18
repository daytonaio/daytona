// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetContainerRegistry godoc
//
//	@Tags			container-registry
//	@Summary		Set container registry credentials
//	@Description	Set container registry credentials
//	@Param			containerRegistry	body	ContainerRegistry	true	"Container Registry credentials to set"
//	@Success		201
//	@Router			/container-registry [put]
//
//	@id				SetContainerRegistry
func SetContainerRegistry(ctx *gin.Context) {
	var req containerregistry.ContainerRegistry
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	cr, err := server.ContainerRegistryService.Find(req.Server, req.Username)
	if err == nil {
		cr.Server = req.Server
		cr.Username = req.Username
		cr.Password = req.Password
	} else {
		cr = &req
	}

	err = server.ContainerRegistryService.Save(cr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set container registry: %s", err.Error()))
		return
	}

	ctx.Status(201)
}
