//go:build linux

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"log/slog"
	"os"
	"path"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	"github.com/daytonaio/daemon/pkg/toolbox/lsp"
	"github.com/daytonaio/daemon/pkg/toolbox/process/interpreter"
	"github.com/daytonaio/daemon/pkg/toolbox/process/pty"
	"github.com/gin-gonic/gin"
)

func (s *server) registerPlatformRoutes(r *gin.Engine) {
	lspLogger := s.logger.With(slog.String("component", "lsp_service"))
	lspController := r.Group("/lsp")
	{
		lspController.POST("/start", lsp.Start(lspLogger))
		lspController.POST("/stop", lsp.Stop(lspLogger))

		lspController.POST("/completions", lsp.Completions(lspLogger))
		lspController.POST("/did-open", lsp.DidOpen(lspLogger))
		lspController.POST("/did-close", lsp.DidClose(lspLogger))

		lspController.GET("/document-symbols", lsp.DocumentSymbols(lspLogger))
		lspController.GET("/workspacesymbols", lsp.WorkspaceSymbols(lspLogger))
	}

	processController := r.Group("/process")
	{
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

		interpreterController := interpreter.NewInterpreterController(s.logger, s.WorkDir)
		interpreterGroup := processController.Group("/interpreter")
		{
			interpreterGroup.POST("/context", interpreterController.CreateContext)
			interpreterGroup.GET("/context", interpreterController.ListContexts)
			interpreterGroup.DELETE("/context/:id", interpreterController.DeleteContext)
			interpreterGroup.GET("/execute", interpreterController.Execute)
		}
	}

	lazyCU := computeruse.NewLazyComputerUse()
	s.computerUse = lazyCU

	go func() {
		pluginPath := "/usr/local/lib/daytona-computer-use"
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			pluginPath = path.Join(s.configDir, "daytona-computer-use")
		}

		if _, err := manager.GetComputerUse(s.logger, pluginPath, lazyCU.Set); err != nil {
			s.logger.Error("Computer-Use error", "error", err)
			s.logger.Info("Continuing without computer-use functionality...")
			return
		}
		s.logger.Info("Computer-use plugin loaded successfully")
	}()

	computerUseController := r.Group("/computeruse")
	computerUseHandler := computeruse.Handler{ComputerUse: lazyCU}

	cuRoutes := s.registerComputerUseRoutes(computerUseController, lazyCU, computerUseHandler)
	cuRoutes.POST("/start", computerUseHandler.StartComputerUse)
	cuRoutes.POST("/stop", computerUseHandler.StopComputerUse)
}

func (s *server) shutdownPlatform() {
	if s.computerUse == nil || !s.computerUse.IsReady() {
		return
	}
	s.logger.Info("Stopping computer use...")
	if _, err := s.computerUse.Stop(); err != nil {
		s.logger.Error("Failed to stop computer use", "error", err)
	}
}
