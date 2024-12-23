// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWorkspaceTemplate godoc
//
//	@Tags			workspace-template
//	@Summary		Get workspace template data
//	@Description	Get workspace template data
//	@Accept			json
//	@Param			templateName	path		string	true	"Template name"
//	@Success		200				{object}	WorkspaceTemplate
//	@Router			/workspace-template/{templateName} [get]
//
//	@id				GetWorkspaceTemplate
func GetWorkspaceTemplate(ctx *gin.Context) {
	templateName := ctx.Param("templateName")

	server := server.GetInstance(nil)

	workspaceTemplate, err := server.WorkspaceTemplateService.Find(ctx.Request.Context(), &stores.WorkspaceTemplateFilter{
		Name: &templateName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace template: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceTemplate)
}

// GetDefaultWorkspaceTemplate godoc
//
//	@Tags			workspace-template
//	@Summary		Get workspace templates by git url
//	@Description	Get workspace templates by git url
//	@Produce		json
//	@Param			gitUrl	path		string	true	"Git URL"
//	@Success		200		{object}	WorkspaceTemplate
//	@Router			/workspace-template/default/{gitUrl} [get]
//
//	@id				GetDefaultWorkspaceTemplate
func GetDefaultWorkspaceTemplate(ctx *gin.Context) {
	gitUrl := ctx.Param("gitUrl")

	decodedURLParam, err := url.QueryUnescape(gitUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to decode query param: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	workspaceTemplates, err := server.WorkspaceTemplateService.Find(ctx.Request.Context(), &stores.WorkspaceTemplateFilter{
		Url:     &decodedURLParam,
		Default: util.Pointer(true),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsWorkspaceTemplateNotFound(err) {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatus(statusCode)
			log.Debugf("Workspace template not added for git url: %s", decodedURLParam)
			return
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find workspace template by git url: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceTemplates)
}

// ListWorkspaceTemplates godoc
//
//	@Tags			workspace-template
//	@Summary		List workspace templates
//	@Description	List workspace templates
//	@Produce		json
//	@Success		200	{array}	WorkspaceTemplate
//	@Router			/workspace-template [get]
//
//	@id				ListWorkspaceTemplates
func ListWorkspaceTemplates(ctx *gin.Context) {
	server := server.GetInstance(nil)

	workspaceTemplates, err := server.WorkspaceTemplateService.List(ctx.Request.Context(), nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspace templates: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceTemplates)
}

// SetWorkspaceTemplate godoc
//
//	@Tags			workspace-template
//	@Summary		Set workspace template data
//	@Description	Set workspace template data
//	@Accept			json
//	@Param			workspaceTemplate	body	CreateWorkspaceTemplateDTO	true	"Workspace template"
//	@Success		201
//	@Router			/workspace-template [put]
//
//	@id				SetWorkspaceTemplate
func SetWorkspaceTemplate(ctx *gin.Context) {
	var req services.CreateWorkspaceTemplateDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	s := server.GetInstance(nil)

	workspaceTemplate := models.WorkspaceTemplate{
		Name:                req.Name,
		BuildConfig:         req.BuildConfig,
		RepositoryUrl:       req.RepositoryUrl,
		EnvVars:             req.EnvVars,
		GitProviderConfigId: req.GitProviderConfigId,
	}

	if req.Image != nil {
		workspaceTemplate.Image = *req.Image
	}

	if req.User != nil {
		workspaceTemplate.User = *req.User
	}

	err = s.WorkspaceTemplateService.Save(ctx.Request.Context(), &workspaceTemplate)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save workspace template: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// SetDefaultWorkspaceTemplate godoc
//
//	@Tags			workspace-template
//	@Summary		Set workspace template to default
//	@Description	Set workspace template to default
//	@Param			templateName	path	string	true	"Template name"
//	@Success		200
//	@Router			/workspace-template/{templateName}/set-default [patch]
//
//	@id				SetDefaultWorkspaceTemplate
func SetDefaultWorkspaceTemplate(ctx *gin.Context) {
	templateName := ctx.Param("templateName")

	server := server.GetInstance(nil)

	err := server.WorkspaceTemplateService.SetDefault(ctx.Request.Context(), templateName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to set workspace template to default: %s", err.Error()))
		return
	}

	ctx.Status(200)
}

// DeleteWorkspaceTemplate godoc
//
//	@Tags			workspace-template
//	@Summary		Delete workspace template data
//	@Description	Delete workspace template data
//	@Param			templateName	path	string	true	"Template name"
//	@Param			force			query	bool	false	"Force"
//	@Success		204
//	@Router			/workspace-template/{templateName} [delete]
//
//	@id				DeleteWorkspaceTemplate
func DeleteWorkspaceTemplate(ctx *gin.Context) {
	templateName := ctx.Param("templateName")
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

	workspaceTemplate, err := server.WorkspaceTemplateService.Find(ctx.Request.Context(), &stores.WorkspaceTemplateFilter{
		Name: &templateName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find workspace template: %s", err.Error()))
		return
	}

	errs := server.WorkspaceTemplateService.Delete(ctx.Request.Context(), workspaceTemplate.Name, force)
	if len(errs) > 0 {
		if stores.IsWorkspaceTemplateNotFound(errs[0]) {
			ctx.AbortWithError(http.StatusNotFound, errors.New("workspace template not found"))
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
