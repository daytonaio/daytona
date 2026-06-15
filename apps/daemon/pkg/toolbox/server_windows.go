//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"net/http"
	"os"
	"path/filepath"

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

	// The manager is the single writer of lazyCU: GetComputerUse and
	// KillComputerUse update it via lazyCU.Set while holding the manager
	// lock, so spawn+publish and kill+clear are each atomic and serialized
	// against one another. A stop (or shutdown) racing an in-flight spawn
	// waits for the spawn to finish and then kills the fresh instance — no
	// code path can spawn twice or leak a child process.
	computerUseController := r.Group("/computeruse")
	{
		computerUseController.POST("/start", func(c *gin.Context) {
			pluginPath := resolvePluginPath(s.configDir)
			if _, err := manager.GetComputerUse(s.logger, pluginPath, lazyCU.Set); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "Failed to spawn computer-use plugin in active console session",
					"details": err.Error(),
					"hint":    "Ensure a user is logged on (AutoLogon) and the plugin binary is at " + pluginPath,
				})
				return
			}
			cuHandler.StartComputerUse(c)
		})

		computerUseController.POST("/stop", func(c *gin.Context) {
			if lazyCU.IsReady() {
				cuHandler.StopComputerUse(c)
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "Computer-use plugin was not running"})
			}
			// Unconditional: waits for and kills a child that a racing
			// /start may still be spawning; cheap no-op when nothing was
			// spawned.
			manager.KillComputerUse(lazyCU.Set)
		})
	}

	s.registerComputerUseRoutes(computerUseController, lazyCU, cuHandler)
}

func (s *server) shutdownPlatform() {
	if s.computerUse == nil {
		// Routes were never registered, so nothing could have spawned.
		return
	}
	if s.computerUse.IsReady() {
		s.logger.Info("Stopping computer-use plugin...")
		if _, err := s.computerUse.Stop(); err != nil {
			s.logger.Error("Failed to stop computer-use plugin", "error", err)
		}
	}
	// NOT gated on IsReady: an in-flight /start spawn (console-session token
	// poll, up to 60s) can outlive the HTTP drain timeout, leaving IsReady
	// false while the manager is creating a live child — and Windows has no
	// parent-death kill, so skipping this would orphan that child in the
	// console session. KillComputerUse takes the manager lock, so it waits
	// for any in-flight spawn, then terminates the fresh instance; it is a
	// no-op when nothing was spawned.
	manager.KillComputerUse(s.computerUse.Set)
}
