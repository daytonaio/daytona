// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/config"
	"github.com/gin-gonic/gin"
)

// GetPrebuild godoc
//
//	@Tags			prebuild
//	@Summary		Get prebuild
//	@Description	Get prebuild
//	@Accept			json
//	@Param			configName	path		string	true	"Workspace config name"
//	@Param			prebuildId	path		string	true	"Prebuild ID"
//	@Success		200			{object}	PrebuildDTO
//	@Router			/workspace-config/{configName}/prebuild/{prebuildId} [get]
//
//	@id				GetPrebuild
func GetPrebuild(ctx *gin.Context) {
	configName := ctx.Param("configName")
	prebuildId := ctx.Param("prebuildId")

	server := server.GetInstance(nil)
	res, err := server.WorkspaceConfigService.FindPrebuild(&config.WorkspaceConfigFilter{
		Name: &configName,
	}, &config.PrebuildFilter{
		Id: &prebuildId,
	})
	if err != nil {
		if config.IsPrebuildNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, errors.New("prebuild not found"))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuild: %s", err.Error()))
		return
	}

	ctx.JSON(200, res)
}

// SetPrebuild godoc

// @Tags			prebuild
// @Summary		Set prebuild
// @Description	Set prebuild
// @Accept			json
// @Param			configName	path		string				true	"Config name"
// @Param			prebuild	body		CreatePrebuildDTO	true	"Prebuild"
// @Success		201			{string}	prebuildId
// @Router			/workspace-config/{configName}/prebuild [put]
//
// @id				SetPrebuild
func SetPrebuild(ctx *gin.Context) {
	configName := ctx.Param("configName")

	var dto dto.CreatePrebuildDTO
	err := ctx.BindJSON(&dto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	prebuild, err := server.WorkspaceConfigService.SetPrebuild(configName, dto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set prebuild: %s", err.Error()))
		return
	}

	ctx.String(201, prebuild.Id)
}

// ListPrebuilds godoc

// @Tags			prebuild
// @Summary		List prebuilds
// @Description	List prebuilds
// @Accept			json
// @Success		200	{array}	PrebuildDTO
// @Router			/workspace-config/prebuild [get]
//
// @id				ListPrebuilds
func ListPrebuilds(ctx *gin.Context) {
	server := server.GetInstance(nil)
	res, err := server.WorkspaceConfigService.ListPrebuilds(nil, nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuilds: %s", err.Error()))
		return
	}

	ctx.JSON(200, res)
}

// ListPrebuildsForWorkspaceConfig godoc

// @Tags			prebuild
// @Summary		List prebuilds for workspace config
// @Description	List prebuilds for workspace config
// @Accept			json
// @Param			configName	path	string	true	"Config name"
// @Success		200			{array}	PrebuildDTO
// @Router			/workspace-config/{configName}/prebuild [get]
//
// @id				ListPrebuildsForWorkspaceConfig
func ListPrebuildsForWorkspaceConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	var workspaceConfigFilter *config.WorkspaceConfigFilter

	if configName != "" {
		workspaceConfigFilter = &config.WorkspaceConfigFilter{
			Name: &configName,
		}
	}

	server := server.GetInstance(nil)
	res, err := server.WorkspaceConfigService.ListPrebuilds(workspaceConfigFilter, nil)
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
//	@Param			configName	path	string	true	"Workspace config name"
//	@Param			prebuildId	path	string	true	"Prebuild ID"
//	@Param			force		query	bool	false	"Force"
//	@Success		204
//	@Router			/workspace-config/{configName}/prebuild/{prebuildId} [delete]
//
//	@id				DeletePrebuild
func DeletePrebuild(ctx *gin.Context) {
	configName := ctx.Param("configName")
	prebuildId := ctx.Param("prebuildId")
	forceQuery := ctx.Query("force")

	var err error
	var force bool

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)
	errs := server.WorkspaceConfigService.DeletePrebuild(configName, prebuildId, force)
	if len(errs) > 0 {
		if config.IsPrebuildNotFound(errs[0]) {
			ctx.AbortWithError(http.StatusNotFound, errors.New("prebuild not found"))
			return
		}
		for _, err := range errs {
			_ = ctx.Error(err)
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(204)
}
