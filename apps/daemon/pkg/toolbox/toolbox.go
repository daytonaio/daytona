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
	ProjectDir string
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
