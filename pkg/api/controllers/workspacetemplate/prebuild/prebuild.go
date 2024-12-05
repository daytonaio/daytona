// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

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

// GetPrebuild godoc
//
//	@Tags			prebuild
//	@Summary		Get prebuild
//	@Description	Get prebuild
//	@Accept			json
//	@Param			templateName	path		string	true	"Workspace template name"
//	@Param			prebuildId		path		string	true	"Prebuild ID"
//	@Success		200				{object}	PrebuildDTO
//	@Router			/workspace-template/{templateName}/prebuild/{prebuildId} [get]
//
//	@id				GetPrebuild
func GetPrebuild(ctx *gin.Context) {
	templateName := ctx.Param("templateName")
	prebuildId := ctx.Param("prebuildId")

	server := server.GetInstance(nil)
	res, err := server.WorkspaceTemplateService.FindPrebuild(ctx.Request.Context(), &stores.WorkspaceTemplateFilter{
		Name: &templateName,
	}, &stores.PrebuildFilter{
		Id: &prebuildId,
	})
	if err != nil {
		if stores.IsPrebuildNotFound(err) {
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
// @Param			templateName	path		string				true	"Template name"
// @Param			prebuild		body		CreatePrebuildDTO	true	"Prebuild"
// @Success		201				{string}	prebuildId
// @Router			/workspace-template/{templateName}/prebuild [put]
//
// @id				SetPrebuild
func SetPrebuild(ctx *gin.Context) {
	templateName := ctx.Param("templateName")

	var dto services.CreatePrebuildDTO
	err := ctx.BindJSON(&dto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	prebuild, err := server.WorkspaceTemplateService.SetPrebuild(ctx.Request.Context(), templateName, dto)
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
// @Router			/workspace-template/prebuild [get]
//
// @id				ListPrebuilds
func ListPrebuilds(ctx *gin.Context) {
	server := server.GetInstance(nil)
	res, err := server.WorkspaceTemplateService.ListPrebuilds(ctx.Request.Context(), nil, nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get prebuilds: %s", err.Error()))
		return
	}

	ctx.JSON(200, res)
}

// ListPrebuildsForWorkspaceTemplate godoc

// @Tags			prebuild
// @Summary		List prebuilds for workspace template
// @Description	List prebuilds for workspace template
// @Accept			json
// @Param			templateName	path	string	true	"Template name"
// @Success		200				{array}	PrebuildDTO
// @Router			/workspace-template/{templateName}/prebuild [get]
//
// @id				ListPrebuildsForWorkspaceTemplate
func ListPrebuildsForWorkspaceTemplate(ctx *gin.Context) {
	templateName := ctx.Param("templateName")

	var workspaceTemplateFilter *stores.WorkspaceTemplateFilter

	if templateName != "" {
		workspaceTemplateFilter = &stores.WorkspaceTemplateFilter{
			Name: &templateName,
		}
	}

	server := server.GetInstance(nil)
	res, err := server.WorkspaceTemplateService.ListPrebuilds(ctx.Request.Context(), workspaceTemplateFilter, nil)
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
//	@Param			templateName	path	string	true	"Workspace template name"
//	@Param			prebuildId		path	string	true	"Prebuild ID"
//	@Param			force			query	bool	false	"Force"
//	@Success		204
//	@Router			/workspace-template/{templateName}/prebuild/{prebuildId} [delete]
//
//	@id				DeletePrebuild
func DeletePrebuild(ctx *gin.Context) {
	templateName := ctx.Param("templateName")
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
	errs := server.WorkspaceTemplateService.DeletePrebuild(ctx.Request.Context(), templateName, prebuildId, force)
	if len(errs) > 0 {
		if stores.IsPrebuildNotFound(errs[0]) {
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
