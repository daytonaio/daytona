// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import "github.com/gin-gonic/gin"

// ProcessExecuteCommand 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Execute command
//	@Description	Execute command synchronously inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string			true	"Workspace ID or Name"
//	@Param			projectId	path		string			true	"Project ID"
//	@Param			params		body		ExecuteRequest	true	"Execute command request"
//	@Success		200			{object}	ExecuteResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/process/execute [post]
//
//	@id				ProcessExecuteCommand
func ProcessExecuteCommand(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}
