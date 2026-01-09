// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"net/http"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/session"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

// GetSessionCommandLogs godoc
//
//	@Summary		Get session command logs
//	@Description	Get logs for a specific command within a session. Supports both HTTP and WebSocket streaming.
//	@Tags			process
//	@Produce		text/plain
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			commandId	path		string	true	"Command ID"
//	@Param			follow		query		boolean	false	"Follow logs in real-time (WebSocket only)"
//	@Success		200			{string}	string	"Log content"
//	@Router			/process/session/{sessionId}/command/{commandId}/logs [get]
//
//	@id				GetSessionCommandLogs
func (s *SessionController) GetSessionCommandLogs(c *gin.Context) {
	sessionId := c.Param("sessionId")
	cmdId := c.Param("commandId")

	sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
	if sdkVersion != "" {
		session.SetUpgraderSubprotocols([]string{"X-Daytona-SDK-Version~" + sdkVersion})
	} else {
		session.SetUpgraderSubprotocols(nil)
	}

	versionComparison, err := util.CompareVersions(sdkVersion, "0.27.0-0")
	if err != nil {
		log.Debug(err)
		versionComparison = util.Pointer(1)
	}

	opts := session.FetchLogsOptions{
		IsCombinedOutput:   session.IsCombinedOutput(sdkVersion, versionComparison, c.Request.Header),
		IsWebsocketUpgrade: c.Request.Header.Get("Upgrade") == "websocket",
		Follow:             c.Query("follow") == "true",
	}

	logBytes, err := s.sessionService.GetSessionCommandLogs(sessionId, cmdId, c.Request, c.Writer, opts)
	if err != nil {
		c.Error(err)
		return
	}

	if logBytes == nil {
		return
	}

	c.String(http.StatusOK, string(logBytes))
}
