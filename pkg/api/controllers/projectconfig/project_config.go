// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/gin-gonic/gin"
)

// GetProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Get project config data
//	@Description	Get project config data
//	@Accept			json
//	@Param			configName	path		string	true	"Config name"
//	@Success		200			{object}	ProjectConfig
//	@Router			/project-config/{configName} [get]
//
//	@id				GetProjectConfig
func GetProjectConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	projectConfig, err := server.ProjectConfigService.Find(configName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get project config: %s", err.Error()))
		return
	}

	ctx.JSON(200, projectConfig)
}

// GetDefaultProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Get project configs by git url
//	@Description	Get project configs by git url
//	@Produce		json
//	@Param			gitUrl	path		string	true	"Git URL"
//	@Success		200		{object}	ProjectConfig
//	@Router			/project-config/default/{gitUrl} [get]
//
//	@id				GetDefaultProjectConfig
func GetDefaultProjectConfig(ctx *gin.Context) {
	gitUrl := ctx.Param("gitUrl")

	decodedURLParam, err := url.QueryUnescape(gitUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	projectConfigs, err := server.ProjectConfigService.FindDefault(decodedURLParam)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if config.IsProjectConfigNotFound(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find project config by git url: %s", err.Error()))
	}

	ctx.JSON(200, projectConfigs)
}

// ListProjectConfigs godoc
//
//	@Tags			project-config
//	@Summary		List project configs
//	@Description	List project configs
//	@Produce		json
//	@Success		200	{array}	ProjectConfig
//	@Router			/project-config [get]
//
//	@id				ListProjectConfigs
func ListProjectConfigs(ctx *gin.Context) {
	server := server.GetInstance(nil)

	projectConfigs, err := server.ProjectConfigService.List()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list project configs: %s", err.Error()))
		return
	}

	ctx.JSON(200, projectConfigs)
}

// SetProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Set project config data
//	@Description	Set project config data
//	@Accept			json
//	@Param			projectConfig	body	dto.CreateProjectConfigDTO	true	"Project config"
//	@Success		201
//	@Router			/project-config [put]
//
//	@id				SetProjectConfig
func SetProjectConfig(ctx *gin.Context) {
	var req dto.CreateProjectConfigDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	s := server.GetInstance(nil)

	projectConfig := s.ProjectConfigService.ToProjectConfig(req)

	err = s.ProjectConfigService.Save(projectConfig)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save project config: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// DeleteProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Delete project config data
//	@Description	Delete project config data
//	@Param			configName	path	string	true	"Config name"
//	@Success		204
//	@Router			/project-config/{configName} [delete]
//
//	@id				DeleteProjectConfig
func DeleteProjectConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	projectConfig, err := server.ProjectConfigService.Find(configName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find project config: %s", err.Error()))
		return
	}

	err = server.ProjectConfigService.Delete(projectConfig.Name)
	if err != nil {
		if config.IsProjectConfigNotFound(err) {
			ctx.Status(204)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get project config: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
