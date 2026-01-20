// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Runner CH API
//	@version		v0.0.0-dev
//	@description	Daytona Runner API - Cloud Hypervisor

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

	"github.com/daytonaio/runner-ch/cmd/runner/config"
	"github.com/daytonaio/runner-ch/pkg/api/controllers"
	"github.com/daytonaio/runner-ch/pkg/api/middlewares"
	"github.com/daytonaio/runner-ch/pkg/runner"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type ApiServerConfig struct {
	ApiPort     int
	TLSCertFile string
	TLSKeyFile  string
	EnableTLS   bool
}

func NewApiServer(cfg ApiServerConfig, r *runner.Runner) *ApiServer {
	controllers.Runner = r
	return &ApiServer{
		apiPort:     cfg.ApiPort,
		tlsCertFile: cfg.TLSCertFile,
		tlsKeyFile:  cfg.TLSKeyFile,
		enableTLS:   cfg.EnableTLS,
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
	_, err := net.Dial("tcp", fmt.Sprintf(":%d", a.apiPort))
	if err == nil {
		return fmt.Errorf("cannot start API server, port %d is already in use", a.apiPort)
	}

	binding.Validator = &DefaultValidator{validate: validator.New()}

	a.router = gin.New()
	a.router.Use(gin.Recovery())

	gin.SetMode(gin.ReleaseMode)
	if config.GetEnvironment() == "development" {
		gin.SetMode(gin.DebugMode)
	}

	a.router.Use(middlewares.LoggingMiddleware())

	// Public routes (no auth required)
	public := a.router.Group("/")
	public.GET("", controllers.HealthCheck)

	// Metrics endpoint
	metricsController := public.Group("/metrics")
	{
		metricsController.GET("", gin.WrapH(promhttp.Handler()))
	}

	// Stats endpoint (public)
	statsController := public.Group("/stats")
	{
		statsController.GET("/memory", controllers.GetMemoryStats)
		statsController.GET("/memory/view", controllers.GetMemoryStatsView)
	}

	// Protected routes (auth required)
	protected := a.router.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	// Runner info
	infoController := protected.Group("/info")
	{
		infoController.GET("", controllers.RunnerInfo)
	}

	// Sandbox management
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
		sandboxController.POST("/:sandboxId/network-settings", controllers.UpdateNetworkSettings)

		// Proxy endpoints for toolbox and port forwarding
		sandboxController.Any("/:sandboxId/toolbox/*path", controllers.ProxyRequest)
		sandboxController.Any("/:sandboxId/proxy/:port/*path", controllers.ProxyToPort)
	}

	// Snapshot management
	snapshotController := protected.Group("/snapshots")
	{
		snapshotController.POST("/pull", controllers.PullSnapshot)
		snapshotController.POST("/push", controllers.PushSnapshot)
		snapshotController.POST("/create", controllers.CreateSnapshot)
		snapshotController.POST("/build", controllers.BuildSnapshot)
		snapshotController.POST("/tag", controllers.TagImage)
		snapshotController.GET("/exists", controllers.SnapshotExists)
		snapshotController.GET("/info", controllers.GetSnapshotInfo)
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

	log.Infof("Starting API server on port %d (TLS: %v)", a.apiPort, a.enableTLS)

	errChan := make(chan error)
	go func() {
		if a.enableTLS {
			errChan <- a.httpServer.ServeTLS(listener, a.tlsCertFile, a.tlsKeyFile)
		} else {
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

// DefaultValidator implements binding.StructValidator
type DefaultValidator struct {
	validate *validator.Validate
}

func (v *DefaultValidator) ValidateStruct(obj interface{}) error {
	return v.validate.Struct(obj)
}

func (v *DefaultValidator) Engine() interface{} {
	return v.validate
}
