//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

func NewMockRestServer(t *testing.T) *httptest.Server {
	router := gin.Default()
	serverController := router.Group("/server")
	{
		serverController.GET("/config", func(ctx *gin.Context) {
			ctx.JSON(200, &server.Config{
				ProvidersDir:      "",
				RegistryUrl:       "",
				Id:                "",
				ServerDownloadUrl: "",
				ApiPort:           3000,
				HeadscalePort:     4000,
				BinariesPath:      "",
				LogFilePath:       "",
			})
		})
		serverController.POST("/network-key", func(ctx *gin.Context) {
			ctx.JSON(200, &server.NetworkKey{Key: "test-key"})
		})
	}

	gitproviderController := router.Group("/gitprovider")
	{
		gitproviderController.GET("/for-url/:url", func(ctx *gin.Context) {
			// This simulates a non-configured git provider
			ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to get git provider for url"))
		})
	}

	server := httptest.NewServer(router)

	return server
}
