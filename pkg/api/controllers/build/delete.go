// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// DeleteAllBuilds godoc
//
//	@Tags			build
//	@Summary		Delete ALL builds
//	@Description	Delete ALL builds
//	@Param			force	query	bool	false	"Force"
//	@Success		204
//	@Router			/build [delete]
//
//	@id				DeleteAllBuilds
func DeleteAllBuilds(ctx *gin.Context) {
	forceQuery := ctx.Query("force")
	var force bool
	var err error

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	errs := server.BuildService.Delete(ctx.Request.Context(), nil, force)
	if len(errs) > 0 {
		for _, err := range errs {
			_ = ctx.Error(err)
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(204)
}

// DeleteBuild godoc
//
//	@Tags			build
//	@Summary		Delete build
//	@Description	Delete build
//	@Param			buildId	path	string	true	"Build ID"
//	@Param			force	query	bool	false	"Force"
//	@Success		204
//	@Router			/build/{buildId} [delete]
//
//	@id				DeleteBuild
func DeleteBuild(ctx *gin.Context) {
	buildId := ctx.Param("buildId")
	forceQuery := ctx.Query("force")
	var force bool
	var err error

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	errs := server.BuildService.Delete(ctx.Request.Context(), &services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			Id: &buildId,
		},
	}, force)
	if len(errs) > 0 {
		for _, err := range errs {
			_ = ctx.Error(err)
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(204)
}

// DeleteBuildsFromPrebuild godoc
//
//	@Tags			build
//	@Summary		Delete builds
//	@Description	Delete builds
//	@Param			prebuildId	path	string	true	"Prebuild ID"
//	@Param			force		query	bool	false	"Force"
//	@Success		204
//	@Router			/build/prebuild/{prebuildId} [delete]
//
//	@id				DeleteBuildsFromPrebuild
func DeleteBuildsFromPrebuild(ctx *gin.Context) {
	prebuildId := ctx.Param("prebuildId")
	forceQuery := ctx.Query("force")
	var force bool
	var err error

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	// Fail if prebuild does not exist
	_, err = server.WorkspaceTemplateService.FindPrebuild(ctx.Request.Context(), nil, &stores.PrebuildFilter{
		Id: &prebuildId,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to find prebuild: %s", err.Error()))
		return
	}

	errs := server.BuildService.Delete(ctx.Request.Context(), &services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			PrebuildIds: &[]string{prebuildId},
		},
	}, force)
	if len(errs) > 0 {
		for _, err := range errs {
			_ = ctx.Error(err)
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(204)
}
