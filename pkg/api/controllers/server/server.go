// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/gin-gonic/gin"
)

// GetConfig 			godoc
//
//	@Tags			server
//	@Summary		Get the server configuration
//	@Description	Get the server configuration
//	@Produce		json
//	@Success		200	{object}	ServerConfig
//	@Router			/server/config [get]
//
//	@id				GetConfig
func GetConfig(ctx *gin.Context) {
	config, err := server.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %w", err))
		return
	}

	ctx.JSON(200, config)
}

// SetConfig 			godoc
//
//	@Tags			server
//	@Summary		Set the server configuration
//	@Description	Set the server configuration
//	@Accept			json
//	@Produce		json
//	@Param			config	body		ServerConfig	true	"Server configuration"
//	@Success		200		{object}	ServerConfig
//	@Router			/server/config [post]
//
//	@id				SetConfig
func SetConfig(ctx *gin.Context) {
	var c server.Config
	err := ctx.BindJSON(&c)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	err = server.Save(c)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save config: %w", err))
		return
	}

	ctx.JSON(200, c)
}

// GenerateNetworkKey 		godoc
//
//	@Tags			server
//	@Summary		Generate a new authentication key
//	@Description	Generate a new authentication key
//	@Produce		json
//	@Success		200	{object}	NetworkKey
//	@Router			/server/network-key [post]
//
//	@id				GenerateNetworkKey
func GenerateNetworkKey(ctx *gin.Context) {
	s := server.GetInstance(nil)

	authKey, err := s.TailscaleServer.CreateAuthKey(headscale.HEADSCALE_USERNAME)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate network key: %w", err))
		return
	}

	ctx.JSON(200, &server.NetworkKey{Key: authKey})
}

// GetServerLogFiles 		godoc
//
//	@Tags			server
//	@Summary		List server log files
//	@Description	List server log files
//	@Produce		json
//	@Success		200	{array}	string
//	@Router			/server/logs [get]
//
//	@id				GetServerLogFiles
func GetServerLogFiles(ctx *gin.Context) {
	server := server.GetInstance(nil)

	logFiles, err := server.GetLogFiles()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, logFiles)
}
