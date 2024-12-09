// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"fmt"
	"net"
	"net/http"

	"github.com/daytonaio/daytona/pkg/agent/toolbox/fs"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/git"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/lsp"
	"github.com/daytonaio/daytona/pkg/agent/toolbox/process"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ProjectDir string
}

const PORT = 2280

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
	r := gin.Default()

	r.GET("/projectdir", s.GetProjectDir)

	//
	// FileSystem
	//

	// read operations
	r.GET("/files", fs.ListFiles)
	r.GET("/files/download", fs.DownloadFile)
	r.GET("/files/find", fs.FindInFiles)
	r.GET("/files/info", fs.GetFileDetails)
	r.GET("/files/search", fs.SearchFiles)

	// create/modify operations
	r.POST("/files", fs.UploadFile)
	r.POST("/files/createfolder", fs.CreateFolder)
	r.POST("/files/move", fs.MoveFile)
	r.POST("/files/permissions", fs.SetFilePermissions)
	r.POST("/files/replace", fs.ReplaceInFiles)
	r.POST("/files/upload", fs.UploadFile)

	// delete operations
	r.DELETE("/files", fs.DeleteFile)

	//
	// Process
	//

	r.POST("/process/execute", process.ExecuteCommand)

	//
	// Git
	//

	r.POST("/git/clone", git.CloneRepository)
	r.GET("/git/status", git.GetStatus)
	r.POST("/git/commit", git.CommitChanges)
	r.POST("/git/push", git.PushChanges)
	r.GET("/git/branches", git.ListBranches)
	r.POST("/git/branches", git.CreateBranch)
	r.GET("/git/history", git.GetCommitHistory)

	//
	// LSP
	//

	//	server process
	r.POST("/lsp/start", lsp.Start)
	r.POST("/lsp/stop", lsp.Stop)

	//	lsp operations
	r.POST("/lsp/didopen", lsp.DidOpen)
	r.POST("/lsp/didclose", lsp.DidClose)
	r.GET("/lsp/documentsymbols", lsp.DocumentSymbols)
	r.GET("/lsp/workspacesymbols", lsp.WorkspaceSymbols)
	r.POST("/lsp/completion", lsp.Completion)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: r,
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	return httpServer.Serve(listener)
}
