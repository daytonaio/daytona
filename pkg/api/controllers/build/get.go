// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetBuild godoc
//
//	@Tags			build
//	@Summary		Get build data
//	@Description	Get build data
//	@Accept			json
//	@Param			buildId	path		string	true	"Build ID"
//	@Success		200		{object}	BuildDTO
//	@Router			/build/{buildId} [get]
//
//	@id				GetBuild
func GetBuild(ctx *gin.Context) {
	buildId := ctx.Param("buildId")

	server := server.GetInstance(nil)

	b, err := server.BuildService.Find(ctx.Request.Context(), &services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			Id: &buildId,
		},
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsBuildNotFound(err) || services.IsBuildDeleted(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find build: %w", err))
		return
	}

	ctx.JSON(200, b)
}
