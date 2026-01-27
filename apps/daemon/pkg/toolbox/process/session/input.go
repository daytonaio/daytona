// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"net/http"

	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// SendInput godoc
//
//	@Summary		Send input to command
//	@Description	Send input data to a running command in a session for interactive execution
//	@Tags			process
//	@Accept			json
//	@Param			sessionId	path	string					true	"Session ID"
//	@Param			commandId	path	string					true	"Command ID"
//	@Param			request		body	SessionSendInputRequest	true	"Input send request"
//	@Success		204
//	@Router			/process/session/{sessionId}/command/{commandId}/input [post]
//
//	@id				SendInput
func (s *SessionController) SendInput(c *gin.Context) {
	sessionId := c.Param("sessionId")
	commandId := c.Param("commandId")

	var request SessionSendInputRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	err := s.sessionService.SendInput(sessionId, commandId, request.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
