// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Toolbox API
//	@version		v0.0.0-dev
//	@description	Daytona Toolbox API

package toolbox

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/daemon/internal"
	"github.com/daytonaio/daemon/pkg/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/config"
	"github.com/daytonaio/daemon/pkg/toolbox/fs"
	"github.com/daytonaio/daemon/pkg/toolbox/git"
	"github.com/daytonaio/daemon/pkg/toolbox/lsp"
	"github.com/daytonaio/daemon/pkg/toolbox/port"
	"github.com/daytonaio/daemon/pkg/toolbox/process/execute"
	"github.com/daytonaio/daemon/pkg/toolbox/process/interpreter"
	"github.com/daytonaio/daemon/pkg/toolbox/process/pty"
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
	Logger                               *slog.Logger
	WorkDir                              string
	ConfigDir                            string
	ComputerUse                          computeruse.IComputerUse
	SandboxId                            string
	OtelEndpoint                         *string
	TerminationGracePeriodSeconds        int
	TerminationCheckIntervalMilliseconds int
	RecordingService                     *recording.RecordingService
}

func NewServer(config ServerConfig) *server {
	return &server{
		logger:                               config.Logger.With(slog.String("component", "toolbox_server")),
		WorkDir:                              config.WorkDir,
		SandboxId:                            config.SandboxId,
		otelEndpoint:                         config.OtelEndpoint,
		telemetry:                            Telemetry{},
		terminationGracePeriodSeconds:        config.TerminationGracePeriodSeconds,
		terminationCheckIntervalMilliseconds: config.TerminationCheckIntervalMilliseconds,
		configDir:                            config.ConfigDir,
		recordingService:                     config.RecordingService,
	}
}

type server struct {
	WorkDir                              string
	ComputerUse                          computeruse.IComputerUse
	SandboxId                            string
	logger                               *slog.Logger
	otelEndpoint                         *string
	authToken                            string
	telemetry                            Telemetry
	terminationGracePeriodSeconds        int
	terminationCheckIntervalMilliseconds int
	configDir                            string
	recordingService                     *recording.RecordingService
}

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	Logger         *otellog.LoggerProvider
}

func (s *server) Start() error {
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

	r.POST("/init", s.Initialize(otelServiceName))

	r.GET("/version", s.GetVersion)

	// keep /project-dir old behavior for backward compatibility
	r.GET("/project-dir", s.GetUserHomeDir)
	r.GET("/user-home-dir", s.GetUserHomeDir)
	r.GET("/work-dir", s.GetWorkDir)

	fsController := r.Group("/files")
	{
		// read operations
		fsController.GET("/", fs.ListFiles)
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
		executeController := execute.NewExecuteController(processLogger, s.terminationGracePeriodSeconds, s.terminationCheckIntervalMilliseconds)
		processController.POST("/execute", executeController.ExecuteCommand)

		sessionController := session.NewSessionController(s.logger, s.configDir, s.WorkDir, s.terminationGracePeriodSeconds, s.terminationCheckIntervalMilliseconds)
		sessionGroup := processController.Group("/session")
		{
			sessionGroup.GET("", sessionController.ListSessions)
			sessionGroup.POST("", sessionController.CreateSession)
			sessionGroup.POST("/:sessionId/exec", sessionController.SessionExecuteCommand)
			sessionGroup.GET("/:sessionId", sessionController.GetSession)
			sessionGroup.DELETE("/:sessionId", sessionController.DeleteSession)
			sessionGroup.GET("/:sessionId/command/:commandId", sessionController.GetSessionCommand)
			sessionGroup.POST("/:sessionId/command/:commandId/input", sessionController.SendInput)
			sessionGroup.GET("/:sessionId/command/:commandId/logs", sessionController.GetSessionCommandLogs)
		}

		// PTY endpoints
		ptyController := pty.NewPTYController(s.logger, s.WorkDir)
		ptyGroup := processController.Group("/pty")
		{
			ptyGroup.GET("", ptyController.ListPTYSessions)
			ptyGroup.POST("", ptyController.CreatePTYSession)
			ptyGroup.GET("/:sessionId", ptyController.GetPTYSession)
			ptyGroup.DELETE("/:sessionId", ptyController.DeletePTYSession)
			ptyGroup.GET("/:sessionId/connect", ptyController.ConnectPTYSession)
			ptyGroup.POST("/:sessionId/resize", ptyController.ResizePTYSession)
		}

		// Interpreter endpoints
		interpreterController := interpreter.NewInterpreterController(s.logger, s.WorkDir)
		interpreterGroup := processController.Group("/interpreter")
		{
			interpreterGroup.POST("/context", interpreterController.CreateContext)
			interpreterGroup.GET("/context", interpreterController.ListContexts)
			interpreterGroup.DELETE("/context/:id", interpreterController.DeleteContext)
			interpreterGroup.GET("/execute", interpreterController.Execute)
		}
	}

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

	lspLogger := s.logger.With(slog.String("component", "lsp_service"))
	lspController := r.Group("/lsp")
	{
		//	server process
		lspController.POST("/start", lsp.Start(lspLogger))
		lspController.POST("/stop", lsp.Stop(lspLogger))

		//	lsp operations
		lspController.POST("/completions", lsp.Completions(lspLogger))
		lspController.POST("/did-open", lsp.DidOpen(lspLogger))
		lspController.POST("/did-close", lsp.DidClose(lspLogger))

		lspController.GET("/document-symbols", lsp.DocumentSymbols(lspLogger))
		lspController.GET("/workspacesymbols", lsp.WorkspaceSymbols(lspLogger))
	}

	// Initialize plugin-based computer use
	pluginPath := "/usr/local/lib/daytona-computer-use"
	// Fallback to local config directory for development
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		pluginPath = path.Join(s.configDir, "daytona-computer-use")
	}
	var err error
	s.ComputerUse, err = manager.GetComputerUse(s.logger, pluginPath)
	if err != nil {
		s.logger.Error("Computer-Use error", "error", err)
		s.logger.Info("Continuing without computer-use functionality...")
	}

	// Always register computer-use endpoints, but handle the case when plugin is nil
	computerUseController := r.Group("/computeruse")
	{
		if s.ComputerUse != nil {
			computerUseHandler := computeruse.Handler{
				ComputerUse: s.ComputerUse,
			}

			// Computer use status endpoint
			computerUseController.GET("/status", computeruse.WrapStatusHandler(s.ComputerUse.GetStatus))

			// Computer use management endpoints
			computerUseController.POST("/start", computerUseHandler.StartComputerUse)
			computerUseController.POST("/stop", computerUseHandler.StopComputerUse)
			computerUseController.GET("/process-status", computerUseHandler.GetComputerUseStatus)
			computerUseController.GET("/process/:processName/status", computerUseHandler.GetProcessStatus)
			computerUseController.POST("/process/:processName/restart", computerUseHandler.RestartProcess)
			computerUseController.GET("/process/:processName/logs", computerUseHandler.GetProcessLogs)
			computerUseController.GET("/process/:processName/errors", computerUseHandler.GetProcessErrors)

			// Screenshot endpoints
			computerUseController.GET("/screenshot", computeruse.WrapScreenshotHandler(s.ComputerUse.TakeScreenshot))
			computerUseController.GET("/screenshot/region", computeruse.WrapRegionScreenshotHandler(s.ComputerUse.TakeRegionScreenshot))
			computerUseController.GET("/screenshot/compressed", computeruse.WrapCompressedScreenshotHandler(s.ComputerUse.TakeCompressedScreenshot))
			computerUseController.GET("/screenshot/region/compressed", computeruse.WrapCompressedRegionScreenshotHandler(s.ComputerUse.TakeCompressedRegionScreenshot))

			// Mouse control endpoints
			computerUseController.GET("/mouse/position", computeruse.WrapMousePositionHandler(s.ComputerUse.GetMousePosition))
			computerUseController.POST("/mouse/move", computeruse.WrapMoveMouseHandler(s.ComputerUse.MoveMouse))
			computerUseController.POST("/mouse/click", computeruse.WrapClickHandler(s.ComputerUse.Click))
			computerUseController.POST("/mouse/drag", computeruse.WrapDragHandler(s.ComputerUse.Drag))
			computerUseController.POST("/mouse/scroll", computeruse.WrapScrollHandler(s.ComputerUse.Scroll))

			// Keyboard control endpoints
			computerUseController.POST("/keyboard/type", computeruse.WrapTypeTextHandler(s.ComputerUse.TypeText))
			computerUseController.POST("/keyboard/key", computeruse.WrapPressKeyHandler(s.ComputerUse.PressKey))
			computerUseController.POST("/keyboard/hotkey", computeruse.WrapPressHotkeyHandler(s.ComputerUse.PressHotkey))

			// Display info endpoints
			computerUseController.GET("/display/info", computeruse.WrapDisplayInfoHandler(s.ComputerUse.GetDisplayInfo))
			computerUseController.GET("/display/windows", computeruse.WrapWindowsHandler(s.ComputerUse.GetWindows))
		} else {
			// Register all endpoints with disabled middleware when plugin is not available
			computerUseController.GET("/status", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/start", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/stop", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/process-status", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/status", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/process/:processName/restart", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/logs", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/errors", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/screenshot", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/region", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/compressed", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/region/compressed", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/mouse/position", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/mouse/move", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/mouse/click", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/mouse/drag", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/mouse/scroll", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/type", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/key", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/hotkey", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/display/info", computeruse.ComputerUseDisabledMiddleware())
			computerUseController.GET("/display/windows", computeruse.ComputerUseDisabledMiddleware())
		}
	}

	// Recording endpoints - always registered, independent of computer-use plugin
	recordingController := recordingcontroller.NewRecordingController(s.recordingService)
	recordingsGroup := computerUseController.Group("/recordings")
	{
		recordingsGroup.POST("/start", recordingController.StartRecording)
		recordingsGroup.POST("/stop", recordingController.StopRecording)
		recordingsGroup.GET("", recordingController.ListRecordings)
		recordingsGroup.GET("/:id", recordingController.GetRecording)
		recordingsGroup.GET("/:id/download", recordingController.DownloadRecording)
		recordingsGroup.DELETE("/:id", recordingController.DeleteRecording)
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

	httpserver := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.TOOLBOX_API_PORT),
		Handler: r,
	}

	// Print to stdout so the runner can know that the daemon is ready
	fmt.Println("Starting toolbox server on port", config.TOOLBOX_API_PORT)

	listener, err := net.Listen("tcp", httpserver.Addr)
	if err != nil {
		return err
	}

	return httpserver.Serve(listener)
}
