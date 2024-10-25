// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/pkg/api/controllers/target/dto"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/gin-gonic/gin"
)

// SetProjectState 			godoc
//
//	@Tags			target
//	@Summary		Set project state
//	@Description	Set project state
//	@Param			targetId	path	string			true	"Target ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			setState	body	SetProjectState	true	"Set State"
//	@Success		200
//	@Router			/target/{targetId}/{projectId}/state [post]
//
//	@id				SetProjectState
func SetProjectState(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	projectId := ctx.Param("projectId")

	var setProjectStateDTO dto.SetProjectState
	err := ctx.BindJSON(&setProjectStateDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, err = server.TargetService.SetProjectState(targetId, projectId, &project.ProjectState{
		Uptime:    setProjectStateDTO.Uptime,
		UpdatedAt: time.Now().Format(time.RFC1123),
		GitStatus: setProjectStateDTO.GitStatus,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop target %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}
