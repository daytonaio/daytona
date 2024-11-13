// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs/dto"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWorkspaceConfig godoc
//
//	@Tags			workspace-config
//	@Summary		Get workspace config data
//	@Description	Get workspace config data
//	@Accept			json
//	@Param			configName	path		string	true	"Config name"
//	@Success		200			{object}	WorkspaceConfig
//	@Router			/workspace-config/{configName} [get]
//
//	@id				GetWorkspaceConfig
func GetWorkspaceConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	workspaceConfig, err := server.WorkspaceConfigService.Find(&stores.WorkspaceConfigFilter{
		Name: &configName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace config: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceConfig)
}

// GetDefaultWorkspaceConfig godoc
//
//	@Tags			workspace-config
//	@Summary		Get workspace configs by git url
//	@Description	Get workspace configs by git url
//	@Produce		json
//	@Param			gitUrl	path		string	true	"Git URL"
//	@Success		200		{object}	WorkspaceConfig
//	@Router			/workspace-config/default/{gitUrl} [get]
//
//	@id				GetDefaultWorkspaceConfig
func GetDefaultWorkspaceConfig(ctx *gin.Context) {
	gitUrl := ctx.Param("gitUrl")

	decodedURLParam, err := url.QueryUnescape(gitUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	workspaceConfigs, err := server.WorkspaceConfigService.Find(&stores.WorkspaceConfigFilter{
		Url:     &decodedURLParam,
		Default: util.Pointer(true),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsWorkspaceConfigNotFound(err) {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatus(statusCode)
			log.Debugf("Workspace config not added for git url: %s", decodedURLParam)
			return
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find workspace config by git url: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceConfigs)
}

// ListWorkspaceConfigs godoc
//
//	@Tags			workspace-config
//	@Summary		List workspace configs
//	@Description	List workspace configs
//	@Produce		json
//	@Success		200	{array}	WorkspaceConfig
//	@Router			/workspace-config [get]
//
//	@id				ListWorkspaceConfigs
func ListWorkspaceConfigs(ctx *gin.Context) {
	server := server.GetInstance(nil)

	workspaceConfigs, err := server.WorkspaceConfigService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspace configs: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceConfigs)
}

// SetWorkspaceConfig godoc
//
//	@Tags			workspace-config
//	@Summary		Set workspace config data
//	@Description	Set workspace config data
//	@Accept			json
//	@Param			workspaceConfig	body	CreateWorkspaceConfigDTO	true	"Workspace config"
//	@Success		201
//	@Router			/workspace-config [put]
//
//	@id				SetWorkspaceConfig
func SetWorkspaceConfig(ctx *gin.Context) {
	var req dto.CreateWorkspaceConfigDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	s := server.GetInstance(nil)

	workspaceConfig := conversion.ToWorkspaceConfig(req)

	err = s.WorkspaceConfigService.Save(workspaceConfig)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save workspace config: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// SetDefaultWorkspaceConfig godoc
//
//	@Tags			workspace-config
//	@Summary		Set workspace config to default
//	@Description	Set workspace config to default
//	@Param			configName	path	string	true	"Config name"
//	@Success		200
//	@Router			/workspace-config/{configName}/set-default [patch]
//
//	@id				SetDefaultWorkspaceConfig
func SetDefaultWorkspaceConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	err := server.WorkspaceConfigService.SetDefault(configName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to set workspace config to default: %s", err.Error()))
		return
	}

	ctx.Status(200)
}

// DeleteWorkspaceConfig godoc
//
//	@Tags			workspace-config
//	@Summary		Delete workspace config data
//	@Description	Delete workspace config data
//	@Param			configName	path	string	true	"Config name"
//	@Param			force		query	bool	false	"Force"
//	@Success		204
//	@Router			/workspace-config/{configName} [delete]
//
//	@id				DeleteWorkspaceConfig
func DeleteWorkspaceConfig(ctx *gin.Context) {
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

	workspaceConfig, err := server.WorkspaceConfigService.Find(&stores.WorkspaceConfigFilter{
		Name: &configName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find workspace config: %s", err.Error()))
		return
	}

	errs := server.WorkspaceConfigService.Delete(workspaceConfig.Name, force)
	if len(errs) > 0 {
		if stores.IsWorkspaceConfigNotFound(errs[0]) {
			ctx.AbortWithError(http.StatusNotFound, errors.New("workspace config not found"))
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
