// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/internal/util"
)

// TerminateSessionCommand godoc
//
//	@Summary		Terminate a session command
//	@Description	Terminate a running command in a session. The daemon handles platform-specific termination.
//	@Tags			process
//	@Param			sessionId	path	string	true	"Session ID"
//	@Param			commandId	path	string	true	"Command ID"
//	@Success		204
//	@Router			/process/session/{sessionId}/command/{commandId}/terminate [post]
//
//	@id				TerminateSessionCommand
func (s *SessionController) TerminateSessionCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")
	commandId := c.Param("commandId")

	if sessionId == util.EntrypointSessionID {
		c.Error(common_errors.NewBadRequestError(errors.New("can't terminate commands in entrypoint session")))
		return
	}

	err := s.sessionService.TerminateCommand(sessionId, commandId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
