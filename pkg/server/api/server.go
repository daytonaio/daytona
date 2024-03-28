// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

//	@title			Daytona Server API
//	@version		0.1.0
//	@description	Daytona Server API

//	@host		localhost:3000
//	@schemes	http
//	@BasePath	/

package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daytona/pkg/server/api/docs"
	"github.com/daytonaio/daytona/pkg/server/api/middlewares"
	"github.com/gin-contrib/cors"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/binary"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/gitprovider"
	log_controller "github.com/daytonaio/daytona/pkg/server/api/controllers/log"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/provider"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/server"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/target"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace"
	"github.com/daytonaio/daytona/pkg/server/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var httpServer *http.Server
var router *gin.Engine

func GetServer() (*http.Server, error) {
	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Description = "Daytona Server API"
	docs.SwaggerInfo.Title = "Daytona Server API"

	if mode, ok := os.LookupEnv("DAYTONA_SERVER_MODE"); ok && mode == "development" {
		router = gin.Default()
		router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
		}))
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	router.Use(middlewares.LoggingMiddleware())

	public := router.Group("/")
	public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	project := protected.Group("/")
	project.Use(middlewares.ProjectAuthMiddleware())

	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	serverController := protected.Group("/server")
	{
		serverController.GET("/config", server.GetConfig)
		serverController.POST("/config", server.SetConfig)
		serverController.POST("/network-key", server.GenerateNetworkKey)
	}

	binaryController := protected.Group("/binary")
	{
		binaryController.GET("/script", binary.GetDaytonaScript)
		binaryController.GET("/:version/:binaryName", binary.GetBinary)
	}

	workspaceController := protected.Group("/workspace")
	{
		workspaceController.GET("/:workspaceId", workspace.GetWorkspace)
		workspaceController.GET("/", workspace.ListWorkspaces)
		workspaceController.POST("/", workspace.CreateWorkspace)
		workspaceController.POST("/:workspaceId/start", workspace.StartWorkspace)
		workspaceController.POST("/:workspaceId/stop", workspace.StopWorkspace)
		workspaceController.DELETE("/:workspaceId", workspace.RemoveWorkspace)
		workspaceController.POST("/:workspaceId/:projectId/start", workspace.StartProject)
		workspaceController.POST("/:workspaceId/:projectId/stop", workspace.StopProject)
	}

	providerController := protected.Group("/provider")
	{
		providerController.POST("/install", provider.InstallProvider)
		providerController.GET("/", provider.ListProviders)
		providerController.POST("/:provider/uninstall", provider.UninstallProvider)
		providerController.GET("/:provider/target-manifest", provider.GetTargetManifest)
	}

	targetController := protected.Group("/target")
	{
		targetController.GET("/", target.ListTargets)
		targetController.PUT("/", target.SetTarget)
		targetController.DELETE("/:target", target.RemoveTarget)
	}

	logController := protected.Group("/log")
	{
		logController.GET("/server", log_controller.ReadServerLog)
		logController.GET("/workspace/:workspaceId", log_controller.ReadWorkspaceLog)
	}

	gitProviderController := protected.Group("/gitprovider")
	{
		gitProviderController.PUT("/", gitprovider.SetGitProvider)
		gitProviderController.DELETE("/:gitProviderId", gitprovider.RemoveGitProvider)
		gitProviderController.GET("/:gitProviderId/user", gitprovider.GetGitUser)
		gitProviderController.GET("/:gitProviderId/namespaces", gitprovider.GetNamespaces)
		gitProviderController.GET("/:gitProviderId/:namespaceId/repositories", gitprovider.GetRepositories)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/branches", gitprovider.GetRepoBranches)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/pull-requests", gitprovider.GetRepoPRs)
		gitProviderController.GET("/context/:gitUrl", gitprovider.GetGitContext)
		gitProviderController.GET("/username-from-token", gitprovider.GetGitUsernameFromToken)
	}

	project.GET(gitProviderController.BasePath()+"/for-url/:url", middlewares.ProjectAuthMiddleware(), gitprovider.GetGitProviderForUrl)

	httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ApiPort),
		Handler: router,
	}

	return httpServer, nil
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
