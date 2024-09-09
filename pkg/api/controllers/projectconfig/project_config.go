// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

	projectConfig, err := server.ProjectConfigService.Find(&config.ProjectConfigFilter{
		Name: &configName,
	})
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

	projectConfigs, err := server.ProjectConfigService.Find(&config.ProjectConfigFilter{
		Url:     &decodedURLParam,
		Default: util.Pointer(true),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if config.IsProjectConfigNotFound(err) {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatus(statusCode)
			log.Debugf("Project config not added for git url: %s", decodedURLParam)
			return
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find project config by git url: %s", err.Error()))
		return
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

	projectConfigs, err := server.ProjectConfigService.List(nil)
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
//	@Param			projectConfig	body	CreateProjectConfigDTO	true	"Project config"
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

	projectConfig := conversion.ToProjectConfig(req)

	err = s.ProjectConfigService.Save(projectConfig)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save project config: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// SetDefaultProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Set project config to default
//	@Description	Set project config to default
//	@Param			configName	path	string	true	"Config name"
//	@Success		200
//	@Router			/project-config/{configName}/set-default [patch]
//
//	@id				SetDefaultProjectConfig
func SetDefaultProjectConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	err := server.ProjectConfigService.SetDefault(configName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to set project config to default: %s", err.Error()))
		return
	}

	ctx.Status(200)
}

// DeleteProjectConfig godoc
//
//	@Tags			project-config
//	@Summary		Delete project config data
//	@Description	Delete project config data
//	@Param			configName	path	string	true	"Config name"
//	@Param			force		query	bool	false	"Force"
//	@Success		204
//	@Router			/project-config/{configName} [delete]
//
//	@id				DeleteProjectConfig
func DeleteProjectConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")
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

	projectConfig, err := server.ProjectConfigService.Find(&config.ProjectConfigFilter{
		Name: &configName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find project config: %s", err.Error()))
		return
	}

	errs := server.ProjectConfigService.Delete(projectConfig.Name, force)
	if len(errs) > 0 {
		if config.IsProjectConfigNotFound(errs[0]) {
			ctx.AbortWithError(http.StatusNotFound, errors.New("project config not found"))
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
