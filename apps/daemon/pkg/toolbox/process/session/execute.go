// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/session"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// SessionExecuteCommand godoc
//
//	@Summary		Execute command in session
//	@Description	Execute a command within an existing shell session
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string					true	"Session ID"
//	@Param			request		body		SessionExecuteRequest	true	"Command execution request"
//	@Success		200			{object}	SessionExecuteResponse
//	@Success		202			{object}	SessionExecuteResponse
//	@Router			/process/session/{sessionId}/exec [post]
//
//	@id				SessionExecuteCommand
func (s *SessionController) SessionExecuteCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")

	var request SessionExecuteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate command is not empty (if not already handled by binding)
	if strings.TrimSpace(request.Command) == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("command cannot be empty"))
		return
	}

	// Handle backward compatibility for "async" field
	if request.Async {
		request.RunAsync = true
	}

	sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
	if sdkVersion != "" {
		session.SetUpgraderSubprotocols([]string{"X-Daytona-SDK-Version~" + sdkVersion})
	} else {
		session.SetUpgraderSubprotocols(nil)
	}

	versionComparison, err := util.CompareVersions(sdkVersion, "0.27.0-0")
	if err != nil {
		log.Error(err)
		versionComparison = util.Pointer(1)
	}

	isCombinedOutput := session.IsCombinedOutput(sdkVersion, versionComparison, c.Request.Header)

	executeResult, err := s.sessionService.Execute(sessionId, request.Command, request.RunAsync, isCombinedOutput)
	if err != nil {
		c.Error(fmt.Errorf("failed to execute command: %w", err))
		return
	}

	if request.RunAsync {
		c.JSON(http.StatusAccepted, &SessionExecuteResponse{
			CommandId: executeResult.CommandId,
		})
		return
	}

	c.JSON(http.StatusOK, &SessionExecuteResponse{
		CommandId: executeResult.CommandId,
		Output:    executeResult.Output,
		Stdout:    executeResult.Stdout,
		Stderr:    executeResult.Stderr,
		ExitCode:  executeResult.ExitCode,
	})
}
