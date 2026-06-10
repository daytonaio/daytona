//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	"github.com/gin-gonic/gin"
)

func resolvePluginPath(configDir string) string {
	const pluginExe = "daytona-computer-use.exe"
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), pluginExe)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	candidate := `C:\ProgramData\Daytona\plugins\` + pluginExe
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(configDir, pluginExe)
}

func (s *server) registerPlatformRoutes(r *gin.Engine) {
	lazyCU := computeruse.NewLazyComputerUse()
	s.computerUse = lazyCU

	cuHandler := computeruse.Handler{ComputerUse: lazyCU}

	// cuMu serializes the management sections of /start and /stop (spawn+set
	// vs stop+kill+clear) so they cannot interleave: a stop issued during an
	// in-flight spawn waits for the spawn to finish, then kills the fresh
	// instance. manager.GetComputerUse / manager.KillComputerUse additionally
	// hold the manager's own lock, so no code path can spawn twice or leak a
	// child process.
	var cuMu sync.Mutex

	computerUseController := r.Group("/computeruse")
	{
		computerUseController.POST("/start", func(c *gin.Context) {
			cuMu.Lock()
			if !lazyCU.IsReady() {
				pluginPath := resolvePluginPath(s.configDir)
				impl, err := manager.GetComputerUse(s.logger, pluginPath)
				if err != nil {
					cuMu.Unlock()
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "Failed to spawn computer-use plugin in active console session",
						"details": err.Error(),
						"hint":    "Ensure a user is logged on (AutoLogon) and the plugin binary is at " + pluginPath,
					})
					return
				}
				lazyCU.Set(impl)
				s.logger.Info("Computer-use plugin spawned into active console session", "path", pluginPath)
			}
			cuMu.Unlock()
			cuHandler.StartComputerUse(c)
		})

		computerUseController.POST("/stop", func(c *gin.Context) {
			cuMu.Lock()
			defer cuMu.Unlock()
			if lazyCU.IsReady() {
				cuHandler.StopComputerUse(c)
				manager.KillComputerUse()
				lazyCU.Set(nil)
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Computer-use plugin was not running"})
		})
	}

	s.registerComputerUseRoutes(computerUseController, lazyCU, cuHandler)
}

func (s *server) shutdownPlatform() {
	if s.computerUse == nil || !s.computerUse.IsReady() {
		return
	}
	s.logger.Info("Stopping computer-use plugin...")
	if _, err := s.computerUse.Stop(); err != nil {
		s.logger.Error("Failed to stop computer-use plugin", "error", err)
	}
	manager.KillComputerUse()
}
