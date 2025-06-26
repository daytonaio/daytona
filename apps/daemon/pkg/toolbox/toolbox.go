// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"

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
	"github.com/daytonaio/daemon/pkg/toolbox/process/session"
	"github.com/daytonaio/daemon/pkg/toolbox/proxy"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	ProjectDir  string
	ComputerUse computeruse.IComputerUse
}

type ProjectDirResponse struct {
	Dir string `json:"dir"`
} // @name ProjectDirResponse

func (s *Server) GetProjectDir(ctx *gin.Context) {
	projectDir := ProjectDirResponse{
		Dir: s.ProjectDir,
	}

	ctx.JSON(http.StatusOK, projectDir)
}

func (s *Server) Start() error {
	// Set Gin to release mode in production
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggingMiddleware())
	r.Use(middlewares.ErrorMiddleware())
	binding.Validator = new(DefaultValidator)

	r.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"version": internal.Version,
		})
	})

	r.GET("/project-dir", s.GetProjectDir)

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir := path.Join(dirname, ".daytona")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("configDir", configDir)

	fsController := r.Group("/files")
	{
		// read operations
		fsController.GET("/", fs.ListFiles)
		fsController.GET("/download", fs.DownloadFile)
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

		sessionController := session.NewSessionController(configDir, s.ProjectDir)
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
			// Computer use status endpoint
			computerUseController.GET("/status", computeruse.WrapStatusHandler(s.ComputerUse.GetStatus))

			// Computer use management endpoints
			computerUseController.POST("/start", s.startComputerUse)
			computerUseController.POST("/stop", s.stopComputerUse)
			computerUseController.GET("/process-status", s.getComputerUseStatus)
			computerUseController.GET("/process/:processName/status", s.getProcessStatus)
			computerUseController.POST("/process/:processName/restart", s.restartProcess)
			computerUseController.GET("/process/:processName/logs", s.getProcessLogs)
			computerUseController.GET("/process/:processName/errors", s.getProcessErrors)

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
			computerUseController.GET("/status", s.computerUseDisabledMiddleware())
			computerUseController.POST("/start", s.computerUseDisabledMiddleware())
			computerUseController.POST("/stop", s.computerUseDisabledMiddleware())
			computerUseController.GET("/process-status", s.computerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/status", s.computerUseDisabledMiddleware())
			computerUseController.POST("/process/:processName/restart", s.computerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/logs", s.computerUseDisabledMiddleware())
			computerUseController.GET("/process/:processName/errors", s.computerUseDisabledMiddleware())
			computerUseController.GET("/screenshot", s.computerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/region", s.computerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/compressed", s.computerUseDisabledMiddleware())
			computerUseController.GET("/screenshot/region/compressed", s.computerUseDisabledMiddleware())
			computerUseController.GET("/mouse/position", s.computerUseDisabledMiddleware())
			computerUseController.POST("/mouse/move", s.computerUseDisabledMiddleware())
			computerUseController.POST("/mouse/click", s.computerUseDisabledMiddleware())
			computerUseController.POST("/mouse/drag", s.computerUseDisabledMiddleware())
			computerUseController.POST("/mouse/scroll", s.computerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/type", s.computerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/key", s.computerUseDisabledMiddleware())
			computerUseController.POST("/keyboard/hotkey", s.computerUseDisabledMiddleware())
			computerUseController.GET("/display/info", s.computerUseDisabledMiddleware())
			computerUseController.GET("/display/windows", s.computerUseDisabledMiddleware())
		}
	}

	portDetector := port.NewPortsDetector()

	portController := r.Group("/port")
	{
		portController.GET("", portDetector.GetPorts)
		portController.GET("/:port/in-use", portDetector.IsPortInUse)
	}

	proxyController := r.Group("/proxy")
	{
		proxyController.Any("/:port/*path", common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget))
	}

	go portDetector.Start(context.Background())

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.TOOLBOX_API_PORT),
		Handler: r,
	}

	// Print to stdout so the runner can know that the daemon is ready
	fmt.Println("Starting toolbox server on port", config.TOOLBOX_API_PORT)

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	return httpServer.Serve(listener)
}

// computerUseDisabledMiddleware returns a middleware that handles requests when computer-use is disabled
func (s *Server) computerUseDisabledMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message":  "Computer-use functionality is not available",
			"details":  "The computer-use plugin failed to initialize due to missing dependencies in the runtime environment.",
			"solution": "Install the required X11 dependencies (x11-apps, xvfb, etc.) to enable computer-use functionality. Check the daemon logs for specific error details.",
		})
		c.Abort()
	}
}

// Computer use management handlers
func (s *Server) startComputerUse(ctx *gin.Context) {
	_, err := s.ComputerUse.Start()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to start computer use",
			"details": err.Error(),
		})
		return
	}

	status, err := s.ComputerUse.GetProcessStatus()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get computer use status",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Computer use processes started successfully",
		"status":  status,
	})
}

func (s *Server) stopComputerUse(ctx *gin.Context) {
	_, err := s.ComputerUse.Stop()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to stop computer use",
			"details": err.Error(),
		})
		return
	}

	status, err := s.ComputerUse.GetProcessStatus()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get computer use status",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Computer use processes stopped successfully",
		"status":  status,
	})
}

func (s *Server) getComputerUseStatus(ctx *gin.Context) {
	status, err := s.ComputerUse.GetProcessStatus()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get computer use status",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

func (s *Server) getProcessStatus(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &computeruse.ProcessRequest{
		ProcessName: processName,
	}
	isRunning, err := s.ComputerUse.IsProcessRunning(req)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get process status",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"processName": processName,
		"running":     isRunning,
	})
}

func (s *Server) restartProcess(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &computeruse.ProcessRequest{
		ProcessName: processName,
	}
	_, err := s.ComputerUse.RestartProcess(req)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("Process %s restarted successfully", processName),
		"processName": processName,
	})
}

func (s *Server) getProcessLogs(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &computeruse.ProcessRequest{
		ProcessName: processName,
	}
	logs, err := s.ComputerUse.GetProcessLogs(req)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"processName": processName,
		"logs":        logs,
	})
}

func (s *Server) getProcessErrors(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &computeruse.ProcessRequest{
		ProcessName: processName,
	}
	errors, err := s.ComputerUse.GetProcessErrors(req)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"processName": processName,
		"errors":      errors,
	})
}
