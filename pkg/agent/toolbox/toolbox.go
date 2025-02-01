// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"fmt"
	"net"
	"net/http"

	"github.com/daytonaio/daytona/pkg/agent/toolbox/config"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/fs"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/git"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/lsp"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/process"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/process/session"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Server struct {
	ConfigDir    string
	WorkspaceDir string
}

type WorkspaceDirResponse struct {
	Dir string `json:"dir"`
} // @name WorkspaceDirResponse

func (s *Server) GetWorkspaceDir(ctx *gin.Context) {
	workspaceDir := WorkspaceDirResponse{
		Dir: s.WorkspaceDir,
	}

	ctx.JSON(200, workspaceDir)
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggingMiddleware())
	binding.Validator = new(api.DefaultValidator)

	r.GET("/workspace-dir", s.GetWorkspaceDir)

	fsController := r.Group("/files")
	{
		// read operations
		fsController.GET("", fs.ListFiles)
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

		// delete operations
		fsController.DELETE("", fs.DeleteFile)
	}

	processController := r.Group("/process")
	{
		processController.POST("/execute", process.ExecuteCommand(s.WorkspaceDir))

		sessionController := processController.Group("/session")
		{
			sessionController.GET("", session.ListSessions)
			sessionController.POST("", session.CreateSession(s.WorkspaceDir, s.ConfigDir))
			sessionController.POST("/:sessionId/exec", session.SessionExecuteCommand(s.ConfigDir))
			sessionController.DELETE("/:sessionId", session.DeleteSession(s.ConfigDir))
			sessionController.GET("/:sessionId/command/:commandId/logs", session.GetSessionCommandLogs(s.ConfigDir))
		}
	}

	gitController := r.Group("/git")
	{
		gitController.GET("/branches", git.ListBranches)
		gitController.GET("/history", git.GetCommitHistory)
		gitController.GET("/status", git.GetStatus)

		gitController.POST("/add", git.AddFiles)
		gitController.POST("/branches", git.CreateBranch)
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

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.TOOLBOX_API_PORT),
		Handler: r,
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	return httpServer.Serve(listener)
}
