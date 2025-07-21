// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Runner API
//	@version		v0.0.0-dev
//	@description	Daytona Runner API

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and an API token.

//	@Security	Bearer

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/controllers"
	"github.com/daytonaio/runner/pkg/api/docs"
	"github.com/daytonaio/runner/pkg/api/middlewares"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ApiServerConfig struct {
	ApiPort     int
	TLSCertFile string
	TLSKeyFile  string
	EnableTLS   bool
}

func NewApiServer(config ApiServerConfig) *ApiServer {
	return &ApiServer{
		apiPort:     config.ApiPort,
		tlsCertFile: config.TLSCertFile,
		tlsKeyFile:  config.TLSKeyFile,
		enableTLS:   config.EnableTLS,
	}
}

type ApiServer struct {
	apiPort     int
	tlsCertFile string
	tlsKeyFile  string
	enableTLS   bool
	httpServer  *http.Server
	router      *gin.Engine
}

func (a *ApiServer) Start() error {
	docs.SwaggerInfo.Description = "Daytona Runner API"
	docs.SwaggerInfo.Title = "Daytona Runner API"
	docs.SwaggerInfo.BasePath = "/"

	_, err := net.Dial("tcp", fmt.Sprintf(":%d", a.apiPort))
	if err == nil {
		return fmt.Errorf("cannot start API server, port %d is already in use", a.apiPort)
	}

	binding.Validator = new(DefaultValidator)

	a.router = gin.New()
	a.router.Use(gin.Recovery())

	gin.SetMode(gin.ReleaseMode)
	if config.GetEnvironment() == "development" {
		gin.SetMode(gin.DebugMode)
	}

	a.router.Use(middlewares.LoggingMiddleware())
	a.router.Use(middlewares.ErrorMiddleware())

	public := a.router.Group("/")
	public.GET("", controllers.HealthCheck)

	if config.GetEnvironment() == "development" {
		public.GET("/api/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	protected := a.router.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	metricsController := public.Group("/metrics")
	{
		metricsController.GET("", gin.WrapH(promhttp.Handler()))
	}

	infoController := protected.Group("/info")
	{
		infoController.GET("", controllers.RunnerInfo)
	}

	sandboxController := protected.Group("/sandboxes")
	{
		sandboxController.POST("", controllers.Create)
		sandboxController.GET("/:sandboxId", controllers.Info)
		sandboxController.POST("/:sandboxId/destroy", controllers.Destroy)
		sandboxController.POST("/:sandboxId/start", controllers.Start)
		sandboxController.POST("/:sandboxId/stop", controllers.Stop)
		sandboxController.POST("/:sandboxId/backup", controllers.CreateBackup)
		sandboxController.POST("/:sandboxId/resize", controllers.Resize)
		sandboxController.DELETE("/:sandboxId", controllers.RemoveDestroyed)

		// Add proxy endpoint within the sandbox controller for toolbox
		// Using Any() to handle all HTTP methods for the toolbox proxy
		sandboxController.Any("/:sandboxId/toolbox/*path", controllers.ProxyRequest)
	}

	snapshotController := protected.Group("/snapshots")
	{
		snapshotController.POST("/pull", controllers.PullSnapshot)
		snapshotController.POST("/build", controllers.BuildSnapshot)
		snapshotController.GET("/exists", controllers.SnapshotExists)
		snapshotController.POST("/remove", controllers.RemoveSnapshot)
		snapshotController.GET("/logs", controllers.GetBuildLogs)
	}

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.apiPort),
		Handler: a.router,
	}

	listener, err := net.Listen("tcp", a.httpServer.Addr)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		if a.enableTLS {
			// Start HTTPS server
			errChan <- a.httpServer.ServeTLS(listener, a.tlsCertFile, a.tlsKeyFile)
		} else {
			// Start HTTP server
			errChan <- a.httpServer.Serve(listener)
		}
	}()

	return <-errChan
}

func (a *ApiServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
