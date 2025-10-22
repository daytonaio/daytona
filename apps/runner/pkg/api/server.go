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
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/pkg/api/controllers"
	"github.com/daytonaio/runner/pkg/api/docs"
	"github.com/daytonaio/runner/pkg/api/middlewares"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/log"
	sloggin "github.com/samber/slog-gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ApiServerConfig struct {
	Logger      *slog.Logger
	ApiPort     int
	ApiToken    string
	TLSCertFile string
	TLSKeyFile  string
	EnableTLS   bool
	LogRequests bool
}

func NewApiServer(config ApiServerConfig) *ApiServer {
	return &ApiServer{
		logger:      config.Logger.With(slog.String("component", "server")),
		apiPort:     config.ApiPort,
		apiToken:    config.ApiToken,
		tlsCertFile: config.TLSCertFile,
		tlsKeyFile:  config.TLSKeyFile,
		enableTLS:   config.EnableTLS,
		logRequests: config.LogRequests,
	}
}

type ApiServer struct {
	logger      *slog.Logger
	apiPort     int
	apiToken    string
	tlsCertFile string
	tlsKeyFile  string
	enableTLS   bool
	httpServer  *http.Server
	router      *gin.Engine
	logRequests bool
}

func (a *ApiServer) Start(ctx context.Context) error {
	docs.SwaggerInfo.Description = "Daytona Runner API"
	docs.SwaggerInfo.Title = "Daytona Runner API"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = internal.Version

	_, err := net.Dial("tcp", fmt.Sprintf(":%d", a.apiPort))
	if err == nil {
		return fmt.Errorf("cannot start API server, port %d is already in use", a.apiPort)
	}

	binding.Validator = new(DefaultValidator)

	gin.DefaultWriter = &log.InfoLogWriter{}
	gin.DefaultErrorWriter = &log.ErrorLogWriter{}

	a.router = gin.New()
	a.router.Use(common_errors.Recovery())

	gin.SetMode(gin.ReleaseMode)
	if config.GetEnvironment() == "development" {
		gin.SetMode(gin.DebugMode)
	}

	if a.logRequests {
		a.router.Use(sloggin.New(a.logger))
	}
	a.router.Use(common_errors.NewErrorMiddleware(common.HandlePossibleDockerError))
	a.router.Use(middlewares.RecoverableErrorsMiddleware())
	a.router.Use(otelgin.Middleware("daytona-runner"))

	public := a.router.Group("/")
	public.GET("", controllers.HealthCheck)

	if config.GetEnvironment() == "development" {
		public.GET("/api/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	protected := a.router.Group("/")
	protected.Use(middlewares.AuthMiddleware(a.apiToken))

	metricsController := protected.Group("/metrics")
	{
		metricsController.GET("", gin.WrapH(promhttp.Handler()))
	}

	infoController := protected.Group("/info")
	{
		infoController.GET("", controllers.RunnerInfo)
	}

	sandboxControllerLogger := a.logger.With(slog.String("component", "sandbox_controller"))
	sandboxController := protected.Group("/sandboxes")
	{
		sandboxController.POST("", controllers.Create)
		sandboxController.GET("/:sandboxId", controllers.Info)
		sandboxController.POST("/:sandboxId/destroy", controllers.Destroy)
		sandboxController.POST("/:sandboxId/start", controllers.Start)
		sandboxController.POST("/:sandboxId/stop", controllers.Stop)
		sandboxController.POST("/:sandboxId/backup", controllers.CreateBackup)
		sandboxController.POST("/:sandboxId/resize", controllers.Resize)
		sandboxController.POST("/:sandboxId/recover", controllers.Recover)
		sandboxController.POST("/:sandboxId/is-recoverable", controllers.IsRecoverable)
		sandboxController.DELETE("/:sandboxId", controllers.RemoveDestroyed)
		sandboxController.POST("/:sandboxId/network-settings", controllers.UpdateNetworkSettings)

		// Add proxy endpoint within the sandbox controller for toolbox
		// Using Any() to handle all HTTP methods for the toolbox proxy
		sandboxController.Any("/:sandboxId/toolbox/*path", controllers.ProxyRequest(sandboxControllerLogger))
	}

	snapshotControllerLogger := a.logger.With(slog.String("component", "snapshot_controller"))
	snapshotController := protected.Group("/snapshots")
	{
		snapshotController.POST("/pull", controllers.PullSnapshot(ctx, snapshotControllerLogger))
		snapshotController.POST("/build", controllers.BuildSnapshot(ctx, snapshotControllerLogger))
		snapshotController.POST("/tag", controllers.TagImage)
		snapshotController.GET("/exists", controllers.SnapshotExists)
		snapshotController.GET("/info", controllers.GetSnapshotInfo)
		snapshotController.POST("/remove", controllers.RemoveSnapshot(snapshotControllerLogger))
		snapshotController.GET("/logs", controllers.GetBuildLogs(snapshotControllerLogger))
		snapshotController.POST("/inspect", controllers.InspectSnapshotInRegistry)
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
		a.logger.Error("Failed to shutdown API server", "error", err)
	}
}
