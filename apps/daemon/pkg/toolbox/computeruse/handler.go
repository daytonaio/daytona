// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	ComputerUse IComputerUse
}

// StartComputerUse godoc
//
//	@Summary		Start computer use processes
//	@Description	Start all computer use processes and return their status
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStartResponse
//	@Router			/computeruse/start [post]
//
//	@id				StartComputerUse
func (h *Handler) StartComputerUse(ctx *gin.Context) {
	_, err := h.ComputerUse.Start()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to start computer use",
			"details": err.Error(),
		})
		return
	}

	status, err := h.ComputerUse.GetProcessStatus()
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

// StopComputerUse godoc
//
//	@Summary		Stop computer use processes
//	@Description	Stop all computer use processes and return their status
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStopResponse
//	@Router			/computeruse/stop [post]
//
//	@id				StopComputerUse
func (h *Handler) StopComputerUse(ctx *gin.Context) {
	_, err := h.ComputerUse.Stop()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to stop computer use",
			"details": err.Error(),
		})
		return
	}

	status, err := h.ComputerUse.GetProcessStatus()
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

// GetComputerUseStatus godoc
//
//	@Summary		Get computer use process status
//	@Description	Get the status of all computer use processes
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStatusResponse
//	@Router			/computeruse/process-status [get]
//
//	@id				GetComputerUseStatus
func (h *Handler) GetComputerUseStatus(ctx *gin.Context) {
	status, err := h.ComputerUse.GetStatus()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get computer use status",
			"details": err.Error(),
		})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "unknown",
		})
		return
	}
	ctx.JSON(http.StatusOK, *status)
}

// GetProcessStatus godoc
//
//	@Summary		Get specific process status
//	@Description	Check if a specific computer use process is running
//	@Tags			computer-use
//	@Produce		json
//	@Param			processName	path		string	true	"Process name to check"
//	@Success		200			{object}	ProcessStatusResponse
//	@Router			/computeruse/process/{processName}/status [get]
//
//	@id				GetProcessStatus
func (h *Handler) GetProcessStatus(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &ProcessRequest{
		ProcessName: processName,
	}
	isRunning, err := h.ComputerUse.IsProcessRunning(req)
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

// RestartProcess godoc
//
//	@Summary		Restart specific process
//	@Description	Restart a specific computer use process
//	@Tags			computer-use
//	@Produce		json
//	@Param			processName	path		string	true	"Process name to restart"
//	@Success		200			{object}	ProcessRestartResponse
//	@Router			/computeruse/process/{processName}/restart [post]
//
//	@id				RestartProcess
func (h *Handler) RestartProcess(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &ProcessRequest{
		ProcessName: processName,
	}
	_, err := h.ComputerUse.RestartProcess(req)

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

// GetProcessLogs godoc
//
//	@Summary		Get process logs
//	@Description	Get logs for a specific computer use process
//	@Tags			computer-use
//	@Produce		json
//	@Param			processName	path		string	true	"Process name to get logs for"
//	@Success		200			{object}	ProcessLogsResponse
//	@Router			/computeruse/process/{processName}/logs [get]
//
//	@id				GetProcessLogs
func (h *Handler) GetProcessLogs(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &ProcessRequest{
		ProcessName: processName,
	}
	logs, err := h.ComputerUse.GetProcessLogs(req)

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

// GetProcessErrors godoc
//
//	@Summary		Get process errors
//	@Description	Get errors for a specific computer use process
//	@Tags			computer-use
//	@Produce		json
//	@Param			processName	path		string	true	"Process name to get errors for"
//	@Success		200			{object}	ProcessErrorsResponse
//	@Router			/computeruse/process/{processName}/errors [get]
//
//	@id				GetProcessErrors
func (h *Handler) GetProcessErrors(ctx *gin.Context) {
	processName := ctx.Param("processName")
	req := &ProcessRequest{
		ProcessName: processName,
	}
	errors, err := h.ComputerUse.GetProcessErrors(req)

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
