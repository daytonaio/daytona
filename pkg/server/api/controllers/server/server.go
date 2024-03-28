// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/types"
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
	config, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	for _, provider := range config.GitProviders {
		provider.Token = strings.Repeat("*", len(provider.Token))
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
	var c types.ServerConfig
	err := ctx.BindJSON(&c)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	err = config.Save(&c)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save config: %s", err.Error()))
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
	authKey, err := headscale.CreateAuthKey()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate network key: %s", err.Error()))
		return
	}

	ctx.JSON(200, &types.NetworkKey{Key: authKey})
}
