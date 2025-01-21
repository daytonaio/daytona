// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import "github.com/gin-gonic/gin"

// CreateSession 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Create exec session
//	@Description	Create exec session inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string					true	"Workspace ID or Name"
//	@Param			params		body	CreateSessionRequest	true	"Create session request"
//	@Success		201
//	@Router			/workspace/{workspaceId}/toolbox/process/session [post]
//
//	@id				CreateSession
func CreateSession(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// SessionExecuteCommand 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Execute command in session
//	@Description	Execute command inside a session inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string					true	"Workspace ID or Name"
//	@Param			sessionId	path		string					true	"Session ID"
//	@Param			params		body		SessionExecuteRequest	true	"Execute command request"
//	@Success		200			{object}	SessionExecuteResponse
//	@Router			/workspace/{workspaceId}/toolbox/process/session/{sessionId}/exec [post]
//
//	@id				SessionExecuteCommand
func SessionExecuteCommand(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// DeleteSession 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Delete session
//	@Description	Delete a session inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			sessionId	path	string	true	"Session ID"
//	@Success		204
//	@Router			/workspace/{workspaceId}/toolbox/process/session/{sessionId} [delete]
//
//	@id				DeleteSession
func DeleteSession(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// ListSessions 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		List sessions
//	@Description	List sessions inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Success		200			{array}	Session
//	@Router			/workspace/{workspaceId}/toolbox/process/session [get]
//
//	@id				ListSessions
func ListSessions(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GetSessionCommandLogs 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get session command logs
//	@Description	Get logs of a command inside a session inside workspace project
//	@Description	Connect with websocket to get a stream of the logs
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			commandId	path		string	true	"Command ID"
//	@Success		200			{string}	string	"command logs"
//	@Router			/workspace/{workspaceId}/toolbox/process/session/{sessionId}/command/{commandId}/logs [get]
//
//	@id				GetSessionCommandLogs
func GetSessionCommandLogs(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}
