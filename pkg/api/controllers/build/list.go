// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// ListBuilds godoc
//
//	@Tags			build
//	@Summary		List builds
//	@Description	List builds
//	@Produce		json
//	@Success		200	{array}	BuildDTO
//	@Router			/build [get]
//
//	@id				ListBuilds
func ListBuilds(ctx *gin.Context) {
	server := server.GetInstance(nil)

	builds, err := server.BuildService.List(ctx.Request.Context(), nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list builds: %s", err.Error()))
		return
	}

	ctx.JSON(200, builds)
}

// ListSuccessfulBuilds godoc
//
//	@Tags			build
//	@Summary		List successful builds for Git repository
//	@Description	List successful builds for Git repository
//	@Produce		json
//	@Param			repoUrl	path	string	true	"Repository URL"
//	@Success		200		{array}	BuildDTO
//	@Router			/build/successful/{repoUrl} [get]
//
//	@id				ListSuccessfulBuilds
func ListSuccessfulBuilds(ctx *gin.Context) {
	urlParam := ctx.Param("url")

	decodedUrl, err := url.QueryUnescape(urlParam)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %w", err))
		return
	}

	server := server.GetInstance(nil)

	builds, err := server.BuildService.List(ctx.Request.Context(), &services.BuildFilter{
		StateNames: &[]models.ResourceStateName{models.ResourceStateNameRunSuccessful},
		StoreFilter: stores.BuildFilter{
			RepositoryUrl: &decodedUrl,
		},
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list successful builds: %s", err.Error()))
		return
	}

	ctx.JSON(200, builds)
}
