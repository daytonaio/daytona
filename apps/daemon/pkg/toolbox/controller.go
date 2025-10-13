// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"net/http"
	"os"

	"github.com/daytonaio/daemon/internal"
	"github.com/gin-gonic/gin"
)

// Initialize godoc
//
//	@Summary		Initialize toolbox server
//	@Description	Set the auth token and initialize telemetry for the toolbox server
//	@Tags			server
//	@Produce		json
//	@Param			request	body		InitializeRequest	true	"Initialization request"
//	@Success		200		{object}	map[string]string
//	@Router			/init [post]
//
//	@id				Initialize
func (s *server) Initialize(otelServiceName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req InitializeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		s.authToken = req.Token

		err := s.initTelemetry(ctx.Request.Context(), otelServiceName)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Auth token set and telemetry initialized successfully",
		})
	}
}

// GetWorkDir godoc
//
//	@Summary		Get working directory
//	@Description	Get the current working directory path. This is default directory used for running commands.
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	WorkDirResponse
//	@Router			/work-dir [get]
//
//	@id				GetWorkDir
func (s *server) GetWorkDir(ctx *gin.Context) {
	workDir := WorkDirResponse{
		Dir: s.WorkDir,
	}

	ctx.JSON(http.StatusOK, workDir)
}

// GetUserHomeDir godoc
//
//	@Summary		Get user home directory
//	@Description	Get the current user home directory path.
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	UserHomeDirResponse
//	@Router			/user-home-dir [get]
//
//	@id				GetUserHomeDir
func (s *server) GetUserHomeDir(ctx *gin.Context) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	userHomeDirResponse := UserHomeDirResponse{
		Dir: userHomeDir,
	}

	ctx.JSON(http.StatusOK, userHomeDirResponse)
}

// GetVersion godoc
//
//	@Summary		Get version
//	@Description	Get the current daemon version
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/version [get]
//
//	@id				GetVersion
func (s *server) GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"version": internal.Version,
	})
}
