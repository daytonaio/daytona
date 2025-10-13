// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Daemon API
//	@version		v0.0.0-dev
//	@description	Daytona Daemon API

package toolbox

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/daemon/internal"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	"github.com/daytonaio/daemon/pkg/toolbox/config"
	"github.com/daytonaio/daemon/pkg/toolbox/fs"
	"github.com/daytonaio/daemon/pkg/toolbox/git"
	"github.com/daytonaio/daemon/pkg/toolbox/lsp"
	"github.com/daytonaio/daemon/pkg/toolbox/middlewares"
	"github.com/daytonaio/daemon/pkg/toolbox/port"
	"github.com/daytonaio/daemon/pkg/toolbox/process"
	"github.com/daytonaio/daemon/pkg/toolbox/process/interpreter"
	"github.com/daytonaio/daemon/pkg/toolbox/process/pty"
	"github.com/daytonaio/daemon/pkg/toolbox/process/session"
	"github.com/daytonaio/daemon/pkg/toolbox/proxy"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/daytonaio/daemon/pkg/toolbox/docs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	WorkDir      string
	SandboxId    string
	OtelEndpoint string
}

func NewServer(config ServerConfig) *server {
	return &server{
		WorkDir:      config.WorkDir,
		SandboxId:    config.SandboxId,
		otelEndpoint: config.OtelEndpoint,
		telemetry:    Telemetry{},
	}
}

type server struct {
	WorkDir      string
	ComputerUse  computeruse.IComputerUse
	SandboxId    string
	otelEndpoint string
	authToken    string
	telemetry    Telemetry
}

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	Logger         *otellog.LoggerProvider
}

var EXCLUDED_TELEMETRY_PATHS = []string{"/proxy"}

func (s *server) Start() error {
	docs.SwaggerInfo.Description = "Daytona Daemon API"
	docs.SwaggerInfo.Title = "Daytona Daemon API"
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
	r.Use(middlewares.LoggingMiddleware())
	noTelemetryRouter.Use(middlewares.LoggingMiddleware())
	r.Use(middlewares.ErrorMiddleware())
	noTelemetryRouter.Use(middlewares.ErrorMiddleware())
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

	dirname, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := path.Join(dirname, ".daytona")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

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

	processController := r.Group("/process")
	{
		processController.POST("/execute", process.ExecuteCommand)

		sessionController := session.NewSessionController(configDir, s.WorkDir)
		sessionGroup := processController.Group("/session")
		{
			sessionGroup.GET("", sessionController.ListSessions)
			sessionGroup.POST("", sessionController.CreateSession)
			sessionGroup.POST("/:sessionId/exec", sessionController.SessionExecuteCommand)
			sessionGroup.GET("/:sessionId", sessionController.GetSession)
			sessionGroup.DELETE("/:sessionId", sessionController.DeleteSession)
			sessionGroup.GET("/:sessionId/command/:commandId", sessionController.GetSessionCommand)
			sessionGroup.GET("/:sessionId/command/:commandId/logs", sessionController.GetSessionCommandLogs)
		}

		// PTY endpoints
		ptyController := pty.NewPTYController(s.WorkDir)
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
		interpreterController := interpreter.NewInterpreterController(s.WorkDir)
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

	lspController := r.Group("/lsp")
	{
		//	server process
		lspController.POST("/start", lsp.Start)
		lspController.POST("/stop", lsp.Stop)

		//	lsp operations
		lspController.POST("/completions", lsp.Completions)
		lspController.POST("/did-open", lsp.DidOpen)
		lspController.POST("/did-close", lsp.DidClose)

		lspController.GET("/document-symbols", lsp.DocumentSymbols)
		lspController.GET("/workspacesymbols", lsp.WorkspaceSymbols)
	}

	// Initialize plugin-based computer use
	pluginPath := "/usr/local/lib/daytona-computer-use"
	// Fallback to local config directory for development
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		pluginPath = path.Join(configDir, "daytona-computer-use")
	}
	s.ComputerUse, err = manager.GetComputerUse(pluginPath)
	if err != nil {
		log.Errorf("Failed to initialize computer-use plugin: %v", err)
		log.Info("Continuing without computer-use functionality...")
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
