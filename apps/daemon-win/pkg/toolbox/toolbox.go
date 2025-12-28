// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//	@title			Daytona Windows Daemon API
//	@version		v0.0.0-dev
//	@description	Daytona Windows Daemon API

package toolbox

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/daytonaio/daemon-win/internal"
	"github.com/daytonaio/daemon-win/pkg/toolbox/config"
	"github.com/daytonaio/daemon-win/pkg/toolbox/fs"
	"github.com/daytonaio/daemon-win/pkg/toolbox/git"
	"github.com/daytonaio/daemon-win/pkg/toolbox/middlewares"
	"github.com/daytonaio/daemon-win/pkg/toolbox/port"
	"github.com/daytonaio/daemon-win/pkg/toolbox/process"
	"github.com/daytonaio/daemon-win/pkg/toolbox/process/session"
	"github.com/daytonaio/daemon-win/pkg/toolbox/proxy"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	WorkDir string
}

type WorkDirResponse struct {
	Dir string `json:"dir"`
} // @name WorkDirResponse

type UserHomeDirResponse struct {
	Dir string `json:"dir"`
} // @name UserHomeDirResponse

// GetWorkDir godoc
//
//	@Summary		Get working directory
//	@Description	Get the current working directory path. This is default directory used for running commands.
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	WorkDirResponse
//	@Router			/work-dir [get]
//
//	@id				GetWorkDir
func (s *Server) GetWorkDir(ctx *gin.Context) {
	workDir := WorkDirResponse{
		Dir: s.WorkDir,
	}

	ctx.JSON(http.StatusOK, workDir)
}

// GetUserHomeDir godoc
//
//	@Summary		Get user home directory
//	@Description	Get the current user home directory path.
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	UserHomeDirResponse
//	@Router			/user-home-dir [get]
//
//	@id				GetUserHomeDir
func (s *Server) GetUserHomeDir(ctx *gin.Context) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	userHomeDirResponse := UserHomeDirResponse{
		Dir: userHomeDir,
	}

	ctx.JSON(http.StatusOK, userHomeDirResponse)
}

// GetVersion godoc
//
//	@Summary		Get version
//	@Description	Get the current daemon version
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/version [get]
//
//	@id				GetVersion
func (s *Server) GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"version": internal.Version,
	})
}

func (s *Server) Start() error {
	// Set Gin to release mode in production
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middlewares.Recovery())
	r.Use(middlewares.LoggingMiddleware())
	r.Use(middlewares.ErrorMiddleware())
	binding.Validator = new(DefaultValidator)

	r.GET("/version", s.GetVersion)

	// keep /project-dir old behavior for backward compatibility
	r.GET("/project-dir", s.GetUserHomeDir)
	r.GET("/user-home-dir", s.GetUserHomeDir)
	r.GET("/work-dir", s.GetWorkDir)

	// Get config directory (Windows: %APPDATA%\daytona)
	dirname, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// On Windows, use AppData\Roaming\daytona
	appDataDir := os.Getenv("APPDATA")
	var configDir string
	if appDataDir != "" {
		configDir = filepath.Join(appDataDir, "daytona")
	} else {
		configDir = filepath.Join(dirname, ".daytona")
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	log.Println("configDir", configDir)

	// File system operations
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

	// Process operations
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
	}

	// Git operations
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

	// Port detection
	portDetector := port.NewPortsDetector()
	portController := r.Group("/port")
	{
		portController.GET("", portDetector.GetPorts)
		portController.GET("/:port/in-use", portDetector.IsPortInUse)
	}

	// Proxy
	proxyController := r.Group("/proxy")
	{
		proxyController.Any("/:port/*path", proxy.ProxyHandler)
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
