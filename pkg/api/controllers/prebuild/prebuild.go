// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/prebuild"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// FindPrebuild godoc
//
//	@Tags			prebuild
//	@Summary		Find prebuild
//	@Description	Find prebuild
//	@Param			key	path		string	true	"Prebuild composite key"
//	@Success		200	{object}	prebuild.Prebuild
//	@Router			/prebuild [get]
//
//	@id				FindPrebuild
func FindPrebuild(ctx *gin.Context) {
	key := ctx.Param("key")

	server := server.GetInstance(nil)
	res, err := server.PrebuildService.Find(key)
	if err != nil {
		if prebuild.IsPrebuildNotFound(err) {
			ctx.JSON(200, &prebuild.Prebuild{})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuild: %s", err.Error()))
		return
	}

	ctx.JSON(200, res)
}

// SetPrebuild godoc

// @Tags			prebuild
// @Summary		Upsert prebuild
// @Description	Upsert prebuild
// @Accept			json
// @Param			prebuild	body	prebuild.Prebuild	true	"Prebuild"
// @Success		201
// @Router			/prebuild [get]
//
// @id				SetPrebuild
func SetPrebuild(ctx *gin.Context) {
	var req prebuild.Prebuild
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	err = server.PrebuildService.Set(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to upsert prebuild: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// ListPrebuilds godoc

// @Tags			prebuild
// @Summary		List prebuilds
// @Description	List prebuilds
// @Accept			json
// @Param			prebuild	body	prebuild.PrebuildFilter	true	"Prebuild Filter"
// @Success		201
// @Router			/prebuild/list [get]
//
// @id				ListPrebuilds
func ListPrebuilds(ctx *gin.Context) {
	var req prebuild.PrebuildFilter
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	res, err := server.PrebuildService.List(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuilds: %s", err.Error()))
		return
	}

	ctx.JSON(200, res)
}

// DeletePrebuild godoc
//
//	@Tags			prebuild
//	@Summary		Delete prebuild
//	@Description	Delete prebuild
//	@Accept			json
//	@Param			prebuild	body	prebuild.Prebuild	true	"Prebuild"
//	@Success		204
//	@Router			/prebuild [delete]
//
//	@id				DeletePrebuild
func DeletePrebuild(ctx *gin.Context) {
	var req prebuild.Prebuild
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	err = server.PrebuildService.Delete(&req)
	if err != nil {
		if prebuild.IsPrebuildNotFound(err) {
			ctx.Status(204)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuild: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
