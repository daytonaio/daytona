// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// CreateBuild godoc
//
//	@Tags			build
//	@Summary		Create a build
//	@Description	Create a build
//	@Accept			json
//	@Param			createBuildDto	body		CreateBuildDTO	true	"Create Build DTO"
//	@Success		201				{string}	buildId
//	@Router			/build [post]
//
//	@id				CreateBuild
func CreateBuild(ctx *gin.Context) {
	var createBuildDto services.CreateBuildDTO
	err := ctx.BindJSON(&createBuildDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	s := server.GetInstance(nil)

	buildId, err := s.BuildService.Create(ctx.Request.Context(), createBuildDto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create build: %s", err.Error()))
		return
	}

	ctx.String(201, buildId)
}
