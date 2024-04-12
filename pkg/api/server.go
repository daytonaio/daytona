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
	"net"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daytona/pkg/api/docs"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	"github.com/gin-contrib/cors"

	"github.com/daytonaio/daytona/pkg/api/controllers/apikey"
	"github.com/daytonaio/daytona/pkg/api/controllers/binary"
	"github.com/daytonaio/daytona/pkg/api/controllers/gitprovider"
	log_controller "github.com/daytonaio/daytona/pkg/api/controllers/log"
	"github.com/daytonaio/daytona/pkg/api/controllers/provider"
	"github.com/daytonaio/daytona/pkg/api/controllers/server"
	"github.com/daytonaio/daytona/pkg/api/controllers/target"
	"github.com/daytonaio/daytona/pkg/api/controllers/workspace"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ApiServerConfig struct {
	ApiPort int
}

func NewApiServer(config ApiServerConfig) *ApiServer {
	return &ApiServer{
		apiPort: config.ApiPort,
	}
}

type ApiServer struct {
	apiPort    int
	httpServer *http.Server
	router     *gin.Engine
}

func (a *ApiServer) Start() error {
	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Description = "Daytona Server API"
	docs.SwaggerInfo.Title = "Daytona Server API"

	if mode, ok := os.LookupEnv("DAYTONA_SERVER_MODE"); ok && mode == "development" {
		a.router = gin.Default()
		a.router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
		}))
	} else {
		gin.SetMode(gin.ReleaseMode)
		a.router = gin.New()
		a.router.Use(gin.Recovery())
	}

	a.router.Use(middlewares.LoggingMiddleware())

	public := a.router.Group("/")
	public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	protected := a.router.Group("/")
	protected.Use(middlewares.AuthMiddleware())

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
		gitProviderController.GET("/", gitprovider.ListGitProviders)
		gitProviderController.PUT("/", gitprovider.SetGitProvider)
		gitProviderController.DELETE("/:gitProviderId", gitprovider.RemoveGitProvider)
		gitProviderController.GET("/:gitProviderId/user", gitprovider.GetGitUser)
		gitProviderController.GET("/:gitProviderId/namespaces", gitprovider.GetNamespaces)
		gitProviderController.GET("/:gitProviderId/:namespaceId/repositories", gitprovider.GetRepositories)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/branches", gitprovider.GetRepoBranches)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/pull-requests", gitprovider.GetRepoPRs)
		gitProviderController.GET("/context/:gitUrl", gitprovider.GetGitContext)
	}

	apiKeyController := protected.Group("/apikey")
	{
		apiKeyController.GET("/", apikey.ListClientApiKeys)
		apiKeyController.POST("/:apiKeyName", apikey.GenerateApiKey)
		apiKeyController.DELETE("/:apiKeyName", apikey.RevokeApiKey)
	}

	projectGroup := protected.Group("/")
	projectGroup.Use(middlewares.ProjectAuthMiddleware())
	{
		projectGroup.POST(workspaceController.BasePath()+"/:workspaceId/:projectId/state", workspace.SetProjectState)
		projectGroup.GET(gitProviderController.BasePath()+"/for-url/:url", gitprovider.GetGitProviderForUrl)
	}

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.apiPort),
		Handler: a.router,
	}

	listener, err := net.Listen("tcp", a.httpServer.Addr)
	if err != nil {
		return err
	}

	log.Infof("Starting api server on port %d", a.apiPort)
	return a.httpServer.Serve(listener)
}

func (a *ApiServer) HealthCheck() error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", a.apiPort))
	if err != nil {
		return fmt.Errorf("API health check timed out")
	}
	defer conn.Close()

	return nil
}

func (a *ApiServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
