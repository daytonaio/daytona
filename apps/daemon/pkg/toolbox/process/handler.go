// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"errors"
	"net/http"
	"strconv"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
)

// ListProcesses godoc
//
//	@Summary		List all running processes
//	@Description	List all tracked running processes across all subsystems (sessions, PTY, interpreter, exec, code_run)
//	@Tags			process
//	@Produce		json
//	@Success		200	{object}	ListProcessesResponse
//	@Router			/process/list [get]
//
//	@id				ListProcesses
func ListProcesses(tracker *ProcessTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		processes := tracker.List()
		filtered := make([]ProcessEntry, 0, len(processes))
		for _, trackedProcess := range processes {
			if trackedProcess.Type == ProcessTypeSession && trackedProcess.ID == util.EntrypointSessionID {
				continue
			}
			if trackedProcess.Internal {
				continue
			}
			filtered = append(filtered, trackedProcess)
		}

		c.JSON(http.StatusOK, ListProcessesResponse{Processes: filtered})
	}
}

// KillProcess godoc
//
//	@Summary		Kill a process by PID
//	@Description	Kill a tracked running process by its OS process ID
//	@Tags			process
//	@Param			pid	path	int	true	"OS Process ID"
//	@Success		200
//	@Router			/process/{pid} [delete]
//
//	@id				KillProcess
func KillProcess(tracker *ProcessTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		pidStr := c.Param("pid")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			c.Error(common_errors.NewBadRequestError(err))
			return
		}

		if err := tracker.Kill(pid); err != nil {
			if errors.Is(err, ErrProcessNotFound) {
				c.Error(common_errors.NewNotFoundError(err))
			} else {
				c.Error(err)
			}
			return
		}

		c.Status(http.StatusOK)
	}
}
