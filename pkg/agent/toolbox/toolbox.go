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
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

	ctx.JSON(200, projectDir)
}

func (s *Server) Start() error {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggingMiddleware())
	binding.Validator = new(api.DefaultValidator)

	r.GET("/project-dir", s.GetProjectDir)

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

		// delete operations
		fsController.DELETE("/", fs.DeleteFile)
	}

	processController := r.Group("/process")
	{
		processController.POST("/execute", process.ExecuteCommand)
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
