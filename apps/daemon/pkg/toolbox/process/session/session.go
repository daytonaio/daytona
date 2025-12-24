// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"net/http"

	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/internal/util"
	log "github.com/sirupsen/logrus"
)

// CreateSession godoc
//
//	@Summary		Create a new session
//	@Description	Create a new shell session for command execution
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateSessionRequest	true	"Session creation request"
//	@Success		201
//	@Router			/process/session [post]
//
//	@id				CreateSession
func (s *SessionController) CreateSession(c *gin.Context) {
	var request CreateSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	// for backward compatibility (only sdk clients before 0.103.X), we use the home directory as the default directory
	sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
	versionComparison, err := util.CompareVersions(sdkVersion, "0.103.0-0")
	if err != nil {
		log.Error(err)
		versionComparison = util.Pointer(1)
	}

	isLegacy := versionComparison != nil && *versionComparison < 0 && sdkVersion != "0.0.0-dev"

	err = s.sessionService.Create(request.SessionId, isLegacy)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusCreated)
}

// DeleteSession godoc
//
//	@Summary		Delete a session
//	@Description	Delete an existing shell session
//	@Tags			process
//	@Param			sessionId	path	string	true	"Session ID"
//	@Success		204
//	@Router			/process/session/{sessionId} [delete]
//
//	@id				DeleteSession
func (s *SessionController) DeleteSession(c *gin.Context) {
	sessionId := c.Param("sessionId")

	err := s.sessionService.Delete(c.Request.Context(), sessionId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSessions godoc
//
//	@Summary		List all sessions
//	@Description	Get a list of all active shell sessions
//	@Tags			process
//	@Produce		json
//	@Success		200	{array}	SessionDTO
//	@Router			/process/session [get]
//
//	@id				ListSessions
func (s *SessionController) ListSessions(c *gin.Context) {
	sessions, err := s.sessionService.List()
	if err != nil {
		c.Error(err)
		return
	}

	sessionDTOs := make([]SessionDTO, 0, len(sessions))
	for _, session := range sessions {
		sessionDTOs = append(sessionDTOs, *SessionToDTO(&session))
	}

	c.JSON(http.StatusOK, sessionDTOs)
}

// GetSession godoc
//
//	@Summary		Get session details
//	@Description	Get details of a specific session including its commands
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	SessionDTO
//	@Router			/process/session/{sessionId} [get]
//
//	@id				GetSession
func (s *SessionController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")

	session, err := s.sessionService.Get(sessionId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, SessionToDTO(session))
}

// GetSessionCommand godoc
//
//	@Summary		Get session command details
//	@Description	Get details of a specific command within a session
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			commandId	path		string	true	"Command ID"
//	@Success		200			{object}	CommandDTO
//	@Router			/process/session/{sessionId}/command/{commandId} [get]
//
//	@id				GetSessionCommand
func (s *SessionController) GetSessionCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")
	cmdId := c.Param("commandId")

	command, err := s.sessionService.GetSessionCommand(sessionId, cmdId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, CommandToDTO(command))
}
