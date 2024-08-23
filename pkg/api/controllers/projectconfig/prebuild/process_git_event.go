// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"fmt"
	"net/http"

	_ "github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ProcessGitEvent 			godoc
//
//	@Tags			prebuild
//	@Summary		ProcessGitEvent
//	@Description	ProcessGitEvent
//	@Param			workspace	body	interface{}	true	"Webhook event"
//	@Success		200
//	@Router			/project-config/prebuild/process-git-event [post]
//
//	@id				ProcessGitEvent
func ProcessGitEvent(ctx *gin.Context) {
	server := server.GetInstance(nil)

	gitProvider, err := server.GitProviderService.GetGitProviderForHttpRequest(ctx.Request)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get git provider for request: %s", err.Error()))
		return
	}

	gitEventData, err := gitProvider.ParseEventData(ctx.Request)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to parse event data: %s", err.Error()))
		return
	}

	if gitEventData == nil {
		return
	}

	err = server.ProjectConfigService.ProcessGitEvent(*gitEventData)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to process git event: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
