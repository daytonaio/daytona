// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"fmt"
	"net/http"

	internal_util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetContainerRegistry 			godoc
//
//	@Tags			container registry
//	@Summary		Get container registry
//	@Description	Get container registry
//	@Produce		json
//	@Param			server	path		string	true	"Container registry server"
//	@Param			workspaceId		query		string	false	"Workspace ID or Name"
//	@Success		200			{object}	ContainerRegistry
//	@Router			/container-registry/{server} [get]
//
//	@id				GetContainerRegistry
func GetContainerRegistry(ctx *gin.Context) {
	serverName := ctx.Param("server")
	workspaceId := ctx.Query("workspaceId")

	var envVars map[string]string
	var err error

	server := server.GetInstance(nil)

	serverEnvVars, err := server.EnvironmentVariableService.Map(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variables: %w", err))
		return
	}

	envVars = serverEnvVars

	if workspaceId != "" {
		w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{})
		if err != nil {
			statusCode := http.StatusInternalServerError
			if stores.IsWorkspaceNotFound(err) || services.IsWorkspaceDeleted(err) {
				statusCode = http.StatusNotFound
			}
			ctx.AbortWithError(statusCode, fmt.Errorf("failed to find workspace: %w", err))
			return
		}

		envVars = internal_util.MergeEnvVars(serverEnvVars, w.EnvVars)
	}

	cr := services.EnvironmentVariables(envVars).FindContainerRegistry(serverName)
	if cr == nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find container registry for server: %s", serverName))
		return
	}

	ctx.JSON(200, cr)
}
