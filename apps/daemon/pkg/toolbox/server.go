// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Toolbox API
//	@version		v0.0.0-dev
//	@description	Daytona Toolbox API
//	@license.name	Apache-2.0
//	@license.url	https://www.apache.org/licenses/LICENSE-2.0

package toolbox

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/common-go/pkg/telemetry"
	"github.com/daytonaio/daemon/internal"
	"github.com/daytonaio/daemon/pkg/recording"
	session_svc "github.com/daytonaio/daemon/pkg/session"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/config"
	"github.com/daytonaio/daemon/pkg/toolbox/fs"
	"github.com/daytonaio/daemon/pkg/toolbox/git"
	"github.com/daytonaio/daemon/pkg/toolbox/port"
	"github.com/daytonaio/daemon/pkg/toolbox/process"
	"github.com/daytonaio/daemon/pkg/toolbox/process/coderun"
	"github.com/daytonaio/daemon/pkg/toolbox/process/session"
	"github.com/daytonaio/daemon/pkg/toolbox/proxy"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/daytonaio/daemon/pkg/toolbox/docs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerConfig struct {
	Logger                *slog.Logger
	WorkDir               string
	ConfigDir             string
	SandboxId             string
	OtelEndpoint          *string
	SessionService        *session_svc.SessionService
	RecordingService      *recording.RecordingService
	OrganizationId        *string
	RegionId              *string
	Snapshot              *string
	EntrypointLogFilePath string
}

func NewServer(config ServerConfig) *server {
	return &server{
		logger:                config.Logger.With(slog.String("component", "toolbox_server")),
		WorkDir:               config.WorkDir,
		SandboxId:             config.SandboxId,
		otelEndpoint:          config.OtelEndpoint,
		telemetry:             Telemetry{},
		sessionService:        config.SessionService,
		configDir:             config.ConfigDir,
		recordingService:      config.RecordingService,
		organizationId:        config.OrganizationId,
		regionId:              config.RegionId,
		snapshot:              config.Snapshot,
		entrypointLogFilePath: config.EntrypointLogFilePath,
	}
}

type server struct {
	WorkDir               string
	SandboxId             string
	logger                *slog.Logger
	otelEndpoint          *string
	authToken             string
	telemetry             Telemetry
	sessionService        *session_svc.SessionService
	configDir             string
	recordingService      *recording.RecordingService
	entrypointLogFilePath string
	entrypointLogCancel   context.CancelFunc
	httpServer            *http.Server
	organizationId        *string
	regionId              *string
	snapshot              *string
	computerUse           *computeruse.LazyComputerUse
	ctx                   context.Context
	cancel                context.CancelFunc
}

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *otellog.LoggerProvider
}

// registerComputerUseRoutes registers the platform-independent computer-use
// and recording routes shared by the Linux and Windows daemons, and returns
// the LazyCheckMiddleware-guarded group. The /start and /stop wiring is
// platform-specific and deliberately asymmetric: Linux attaches them inside
// the returned guarded group, which is safe only because its plugin spawns
// at boot. Windows MUST attach them to the unguarded parent controller —
// /start is the only code path that spawns the plugin there, and the guard
// returns 503 until the plugin is ready, so guarding /start would deadlock
// it into a permanent 503. The route set is part of the frozen wire contract
// consumed by the generated SDKs; keep it identical on both platforms.
func (s *server) registerComputerUseRoutes(controller *gin.RouterGroup, lazyCU *computeruse.LazyComputerUse, cuHandler computeruse.Handler) *gin.RouterGroup {
	cuRoutes := controller.Group("/", computeruse.LazyCheckMiddleware(lazyCU))

	cuRoutes.GET("/status", computeruse.WrapStatusHandler(lazyCU.GetStatus))
	cuRoutes.GET("/process-status", cuHandler.GetComputerUseStatus)
	cuRoutes.GET("/process/:processName/status", cuHandler.GetProcessStatus)
	cuRoutes.POST("/process/:processName/restart", cuHandler.RestartProcess)
	cuRoutes.GET("/process/:processName/logs", cuHandler.GetProcessLogs)
	cuRoutes.GET("/process/:processName/errors", cuHandler.GetProcessErrors)

	cuRoutes.GET("/screenshot", computeruse.WrapScreenshotHandler(lazyCU.TakeScreenshot))
	cuRoutes.GET("/screenshot/region", computeruse.WrapRegionScreenshotHandler(lazyCU.TakeRegionScreenshot))
	cuRoutes.GET("/screenshot/compressed", computeruse.WrapCompressedScreenshotHandler(lazyCU.TakeCompressedScreenshot))
	cuRoutes.GET("/screenshot/region/compressed", computeruse.WrapCompressedRegionScreenshotHandler(lazyCU.TakeCompressedRegionScreenshot))

	cuRoutes.GET("/mouse/position", computeruse.WrapMousePositionHandler(lazyCU.GetMousePosition))
	cuRoutes.POST("/mouse/move", computeruse.WrapMoveMouseHandler(lazyCU.MoveMouse))
	cuRoutes.POST("/mouse/click", computeruse.WrapClickHandler(lazyCU.Click))
	cuRoutes.POST("/mouse/drag", computeruse.WrapDragHandler(lazyCU.Drag))
	cuRoutes.POST("/mouse/scroll", computeruse.WrapScrollHandler(lazyCU.Scroll))

	cuRoutes.POST("/keyboard/type", computeruse.WrapTypeTextHandler(lazyCU.TypeText))
	cuRoutes.POST("/keyboard/key", computeruse.WrapPressKeyHandler(lazyCU.PressKey))
	cuRoutes.POST("/keyboard/hotkey", computeruse.WrapPressHotkeyHandler(lazyCU.PressHotkey))

	cuRoutes.GET("/display/info", computeruse.WrapDisplayInfoHandler(lazyCU.GetDisplayInfo))
	cuRoutes.GET("/display/windows", computeruse.WrapWindowsHandler(lazyCU.GetWindows))

	cuRoutes.GET("/a11y/tree", computeruse.WrapGetAccessibilityTreeHandler(lazyCU.GetAccessibilityTree))
	cuRoutes.POST("/a11y/find", computeruse.WrapFindAccessibilityNodesHandler(lazyCU.FindAccessibilityNodes))
	cuRoutes.POST("/a11y/node/focus", computeruse.WrapFocusAccessibilityNodeHandler(lazyCU.FocusAccessibilityNode))
	cuRoutes.POST("/a11y/node/invoke", computeruse.WrapInvokeAccessibilityNodeHandler(lazyCU.InvokeAccessibilityNode))
	cuRoutes.POST("/a11y/node/value", computeruse.WrapSetAccessibilityNodeValueHandler(lazyCU.SetAccessibilityNodeValue))

	recordingController := recordingcontroller.NewRecordingController(s.recordingService)
	recordingsGroup := controller.Group("/recordings")
	{
		recordingsGroup.POST("/start", recordingController.StartRecording)
		recordingsGroup.POST("/stop", recordingController.StopRecording)
		recordingsGroup.GET("", recordingController.ListRecordings)
		recordingsGroup.GET("/:id", recordingController.GetRecording)
		recordingsGroup.GET("/:id/download", recordingController.DownloadRecording)
		recordingsGroup.DELETE("/:id", recordingController.DeleteRecording)
	}

	return cuRoutes
}

func (s *server) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	defer s.cancel()

	docs.SwaggerInfo.Description = "Daytona Toolbox API"
	docs.SwaggerInfo.Title = "Daytona Toolbox API"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = internal.Version

	// Set Gin to release mode in production
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	otelServiceName := fmt.Sprintf("sandbox-%s", s.SandboxId)

	r := gin.New()
	r.Use(common_errors.Recovery())
	noTelemetryRouter := r.Group("/")
	r.Use(func(ctx *gin.Context) {
		if s.telemetry.TracerProvider == nil {
			ctx.Next()
			return
		}

		otelgin.Middleware(otelServiceName, otelgin.WithTracerProvider(s.telemetry.TracerProvider))(ctx)
		ctx.Next()
	})
	r.Use(sloggin.New(s.logger))
	errMiddleware := common_errors.NewErrorMiddleware(func(ctx *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	})

	noTelemetryRouter.Use(sloggin.New(s.logger))
	r.Use(errMiddleware)
	noTelemetryRouter.Use(errMiddleware)
	binding.Validator = new(DefaultValidator)

	// Add swagger UI in development mode
	if os.Getenv("ENVIRONMENT") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	r.POST("/init", s.Initialize(otelServiceName, s.entrypointLogFilePath, s.organizationId, s.regionId, s.snapshot))

	r.GET("/version", s.GetVersion)

	// keep /project-dir old behavior for backward compatibility
	r.GET("/project-dir", s.GetUserHomeDir)
	r.GET("/user-home-dir", s.GetUserHomeDir)
	r.GET("/work-dir", s.GetWorkDir)

	fsController := r.Group("/files")
	{
		// read operations
		fsController.GET("/", fs.ListFiles)
		fsController.GET("", fs.ListFiles)
		fsController.GET("/download", fs.DownloadFile)
		fsController.POST("/bulk-download", fs.DownloadFiles)
		fsController.GET("/find", fs.FindInFiles)
		fsController.GET("/info", fs.GetFileInfo)
		fsController.GET("/search", fs.SearchFiles)

		// create/modify operations
		fsController.POST("/folder", fs.CreateFolder)
		fsController.POST("/move", fs.MoveFile)
		fsController.POST("/permissions", fs.SetFilePermissions)
		fsController.POST("/replace", fs.ReplaceInFiles)
		fsController.POST("/upload", fs.UploadFile)
		fsController.POST("/bulk-upload", fs.UploadFiles)

		// delete operations
		fsController.DELETE("/", fs.DeleteFile)
	}

	processLogger := s.logger.With(slog.String("component", "process_controller"))
	processController := r.Group("/process")
	{
		processController.POST("/execute", process.ExecuteCommand(processLogger))
		processController.POST("/code-run", coderun.CodeRun(processLogger))

		sessionController := session.NewSessionController(s.logger, s.configDir, s.sessionService)
		sessionGroup := processController.Group("/session")
		{
			sessionGroup.GET("", sessionController.ListSessions)
			sessionGroup.POST("", sessionController.CreateSession)
			sessionGroup.GET("/entrypoint", sessionController.GetEntrypointSession)
			sessionGroup.GET("/entrypoint/logs", sessionController.GetEntrypointLogs)
			sessionGroup.POST("/:sessionId/exec", sessionController.SessionExecuteCommand)
			sessionGroup.GET("/:sessionId", sessionController.GetSession)
			sessionGroup.DELETE("/:sessionId", sessionController.DeleteSession)
			sessionGroup.GET("/:sessionId/command/:commandId", sessionController.GetSessionCommand)
			sessionGroup.POST("/:sessionId/command/:commandId/input", sessionController.SendInput)
			sessionGroup.GET("/:sessionId/command/:commandId/logs", sessionController.GetSessionCommandLogs)
		}
	}

	s.registerPlatformRoutes(r)

	gitController := r.Group("/git")
	{
		gitController.GET("/branches", git.ListBranches)
		gitController.GET("/history", git.GetCommitHistory)
		gitController.GET("/status", git.GetStatus)

		gitController.POST("/add", git.AddFiles)
		gitController.POST("/branches", git.CreateBranch)
		gitController.POST("/checkout", git.CheckoutBranch)
		gitController.DELETE("/branches", git.DeleteBranch)
		gitController.POST("/clone", git.CloneRepository)
		gitController.POST("/commit", git.CommitChanges)
		gitController.POST("/pull", git.PullChanges)
		gitController.POST("/push", git.PushChanges)
	}

	portDetector := port.NewPortsDetector()

	portController := r.Group("/port")
	{
		portController.GET("", portDetector.GetPorts)
		portController.GET("/:port/in-use", portDetector.IsPortInUse)
	}

	proxyController := noTelemetryRouter.Group("/proxy")
	{
		proxyController.Any("/:port/*path", common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget, nil))
	}

	go portDetector.Start(context.Background())

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.TOOLBOX_API_PORT),
		Handler: r,
	}

	// Print to stdout so the runner can know that the daemon is ready
	fmt.Println("Starting toolbox server on port", config.TOOLBOX_API_PORT)

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	return s.httpServer.Serve(listener)
}

func (s *server) Shutdown() {
	s.logger.Info("Shutting down toolbox server")

	// Stop accepting new requests and drain in-flight ones
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("toolbox HTTP server shutdown error", "error", err)
		}
	}

	s.shutdownPlatform()

	// Flush telemetry
	if s.telemetry.TracerProvider != nil {
		s.logger.Info("Shutting down tracer provider")
		telemetry.ShutdownTracer(s.logger, s.telemetry.TracerProvider)
	}

	if s.telemetry.MeterProvider != nil {
		s.logger.Info("Shutting down meter provider")
		telemetry.ShutdownMeter(s.logger, s.telemetry.MeterProvider)
	}

	if s.telemetry.LoggerProvider != nil {
		s.logger.Info("Shutting down logger provider")
		telemetry.ShutdownLogger(s.logger, s.telemetry.LoggerProvider)
	}
}
