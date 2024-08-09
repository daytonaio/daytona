// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListBuilds godoc
//
//	@Tags			build
//	@Summary		List builds
//	@Description	List builds
//	@Produce		json
//	@Success		200	{array}	Build
//	@Router			/build [get]
//
//	@id				ListBuilds
func ListBuilds(ctx *gin.Context) {
	server := server.GetInstance(nil)

	builds, err := server.BuildService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list builds: %s", err.Error()))
		return
	}

	ctx.JSON(200, builds)
}
