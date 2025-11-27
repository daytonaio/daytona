// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// Create 			godoc
//
//	@Tags			sandbox
//	@Summary		Create a sandbox
//	@Description	Create a sandbox
//	@Param			sandbox	body	dto.CreateSandboxDTO	true	"Create sandbox"
//	@Produce		json
//	@Success		201	{string}	containerId
//	@Failure		400	{object}	common_errors.ErrorResponse
//	@Failure		401	{object}	common_errors.ErrorResponse
//	@Failure		404	{object}	common_errors.ErrorResponse
//	@Failure		409	{object}	common_errors.ErrorResponse
//	@Failure		500	{object}	common_errors.ErrorResponse
//	@Router			/sandboxes [post]
//
//	@id				Create
func Create(ctx *gin.Context) {
	var createSandboxDto dto.CreateSandboxDTO
	err := ctx.ShouldBindJSON(&createSandboxDto)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	containerId, err := runner.Docker.Create(ctx.Request.Context(), createSandboxDto)
	if err != nil {
		runner.StatesCache.SetSandboxState(ctx, createSandboxDto.Id, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		ctx.Error(err)
		return
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	ctx.JSON(http.StatusCreated, containerId)
}

// Destroy 			godoc
//
//	@Tags			sandbox
//	@Summary		Destroy sandbox
//	@Description	Destroy sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox destroyed"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/destroy [post]
//
//	@id				Destroy
func Destroy(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.Docker.Destroy(ctx.Request.Context(), sandboxId)
	if err != nil {
		runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		ctx.Error(err)
		return
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	ctx.JSON(http.StatusOK, "Sandbox destroyed")
}

// CreateBackup godoc
//
//	@Tags			sandbox
//	@Summary		Create sandbox backup
//	@Description	Create sandbox backup
//	@Produce		json
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			sandbox		body		dto.CreateBackupDTO	true	"Create backup"
//	@Success		201			{string}	string				"Backup started"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/backup [post]
//
//	@id				CreateBackup
func CreateBackup(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var createBackupDTO dto.CreateBackupDTO
	err := ctx.ShouldBindJSON(&createBackupDTO)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.StartBackupCreate(ctx.Request.Context(), sandboxId, createBackupDTO)
	if err != nil {
		runner.StatesCache.SetBackupState(ctx, sandboxId, enums.BackupStateFailed, err)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, "Backup started")
}

// Resize 			godoc
//
//	@Tags			sandbox
//	@Summary		Resize sandbox
//	@Description	Resize sandbox
//	@Produce		json
//	@Param			sandboxId	path		string					true	"Sandbox ID"
//	@Param			sandbox		body		dto.ResizeSandboxDTO	true	"Resize sandbox"
//	@Success		200			{string}	string					"Sandbox resized"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/resize [post]
//
//	@id				Resize
func Resize(ctx *gin.Context) {
	var resizeDto dto.ResizeSandboxDTO
	err := ctx.ShouldBindJSON(&resizeDto)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err = runner.Docker.Resize(ctx.Request.Context(), sandboxId, resizeDto)
	if err != nil {
		runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox resized")
}

// UpdateNetworkSettings godoc
//
//	@Tags			sandbox
//	@Summary		Update sandbox network settings
//	@Description	Update sandbox network settings
//	@Produce		json
//	@Param			sandboxId	path		string							true	"Sandbox ID"
//	@Param			sandbox		body		dto.UpdateNetworkSettingsDTO	true	"Update network settings"
//	@Success		200			{string}	string							"Network settings updated"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/network-settings [post]
//
//	@id				UpdateNetworkSettings
func UpdateNetworkSettings(ctx *gin.Context) {
	var updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO
	err := ctx.ShouldBindJSON(&updateNetworkSettingsDto)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	sandboxId := ctx.Param("sandboxId")
	runner := runner.GetInstance(nil)

	info, err := runner.Docker.ContainerInspect(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}
	containerShortId := info.ID[:12]

	ipAddress := common.GetContainerIpAddress(ctx, info)

	// Return error if container does not have an IP address
	if ipAddress == "" {
		ctx.Error(common_errors.NewInvalidBodyRequestError(errors.New("sandbox does not have an IP address")))
		return
	}

	if updateNetworkSettingsDto.NetworkBlockAll != nil && *updateNetworkSettingsDto.NetworkBlockAll {
		err = runner.NetRulesManager.SetNetworkRules(containerShortId, ipAddress, "")
		if err != nil {
			ctx.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}
	} else if updateNetworkSettingsDto.NetworkAllowList != nil {
		err = runner.NetRulesManager.SetNetworkRules(containerShortId, ipAddress, *updateNetworkSettingsDto.NetworkAllowList)
		if err != nil {
			ctx.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}
	}

	if updateNetworkSettingsDto.NetworkLimitEgress != nil && *updateNetworkSettingsDto.NetworkLimitEgress {
		err = runner.NetRulesManager.SetNetworkLimiter(containerShortId, ipAddress)
		if err != nil {
			ctx.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, "Network settings updated")
}

// GetNetworkSettings godoc
//
//	@Tags			sandbox
//	@Summary		Get sandbox network settings
//	@Description	Get sandbox network settings
//	@Produce		json
//	@Param			sandboxId	path		string							true	"Sandbox ID"
//	@Success		200			{object}	dto.UpdateNetworkSettingsDTO	"Network settings"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/network-settings [get]
//
//	@id				GetNetworkSettings
func GetNetworkSettings(ctx *gin.Context) {
	// TODO: Implement GetNetworkSettings in Docker client
	// sandboxId := ctx.Param("sandboxId")
	// runner := runner.GetInstance(nil)
	// networkSettings, err := runner.Docker.GetNetworkSettings(ctx.Request.Context(), sandboxId)
	// if err != nil {
	// 	ctx.Error(err)
	// 	return
	// }

	// For now, return empty settings
	networkSettings := dto.UpdateNetworkSettingsDTO{
		NetworkBlockAll:  nil,
		NetworkAllowList: nil,
	}

	ctx.JSON(http.StatusOK, networkSettings)
}

// Start 			godoc
//
//	@Tags			sandbox
//	@Summary		Start sandbox
//	@Description	Start sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Param			metadata	body		object	false	"Metadata"
//	@Success		200			{string}	string	"Sandbox started"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/start [post]
//
//	@id				Start
func Start(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	var metadata map[string]string
	err := ctx.ShouldBindJSON(&metadata)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	err = runner.Docker.Start(ctx.Request.Context(), sandboxId, metadata)

	if err != nil {
		runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox started")
}

// Stop 			godoc
//
//	@Tags			sandbox
//	@Summary		Stop sandbox
//	@Description	Stop sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox stopped"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/stop [post]
//
//	@id				Stop
func Stop(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.Docker.Stop(ctx.Request.Context(), sandboxId)
	if err != nil {
		runner.StatesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox stopped")
}

// Info godoc
//
//	@Tags			sandbox
//	@Summary		Get sandbox info
//	@Description	Get sandbox info
//	@Produce		json
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Success		200			{object}	SandboxInfoResponse	"Sandbox info"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId} [get]
//
//	@id				Info
func Info(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	info := runner.SandboxService.GetSandboxStatesInfo(ctx.Request.Context(), sandboxId)

	ctx.JSON(http.StatusOK, SandboxInfoResponse{
		State:       info.SandboxState,
		BackupState: info.BackupState,
		BackupError: info.BackupErrorReason,
	})
}

type SandboxInfoResponse struct {
	State       enums.SandboxState `json:"state"`
	BackupState enums.BackupState  `json:"backupState"`
	BackupError *string            `json:"backupError,omitempty"`
} //	@name	SandboxInfoResponse

// RemoveDestroyed godoc
//
//	@Tags			sandbox
//	@Summary		Remove a destroyed sandbox
//	@Description	Remove a sandbox that has been previously destroyed
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox removed"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId} [delete]
//
//	@id				RemoveDestroyed
func RemoveDestroyed(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.SandboxService.RemoveDestroyedSandbox(ctx.Request.Context(), sandboxId)
	if err != nil {
		if !common_errors.IsNotFoundError(err) {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, "Sandbox removed")
}

// Logs 			godoc
//
//	@Tags			sandbox
//	@Summary		Get sandbox logs
//	@Description	Get the entire log output of a sandbox container
//	@Produce		text/plain
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Param			timestamps	query		boolean	false	"Whether to include timestamps in the logs"
//	@Success		200			{string}	string	"Container logs"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/logs [get]
//
//	@id				Logs
func Logs(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	includeTimestamps := ctx.Query("timestamps") == "true"

	runner := runner.GetInstance(nil)

	// Get container logs using Docker API
	logs, err := runner.Docker.ApiClient().ContainerLogs(ctx.Request.Context(), sandboxId, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: includeTimestamps,
		Tail:       "1000",
	})
	if err != nil {
		ctx.Error(common_errors.NewNotFoundError(errors.New("container not found or logs unavailable")))
		return
	}
	defer logs.Close()

	ctx.Header("Content-Type", "text/plain")

	// Process and normalize logs before sending
	normalizedLogs := normalizeLogs(logs)
	_, err = ctx.Writer.Write([]byte(normalizedLogs))
	if err != nil {
		ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		return
	}
}

func normalizeLogs(logs io.ReadCloser) string {
	var result strings.Builder
	scanner := bufio.NewScanner(logs)

	for scanner.Scan() {
		line := scanner.Text()

		// Handle Docker log format - Docker logs have 8-byte headers
		// First byte indicates stream (1=stdout, 2=stderr), next 3 bytes are unused, last 4 bytes are size
		if len(line) >= 8 {
			// Check if this looks like a Docker log header
			if line[0] == 1 || line[0] == 2 {
				// Skip the 8-byte header and process the actual log content
				if len(line) > 8 {
					line = line[8:]
				} else {
					continue // Skip lines that are just headers
				}
			}
		}

		// Split on embedded newlines and process each part
		parts := strings.Split(line, "\n")
		for i, part := range parts {
			// Skip empty parts
			if strings.TrimSpace(part) == "" {
				continue
			}

			// Normalize whitespace - replace tabs with spaces and normalize indentation
			normalized := normalizeWhitespace(part)
			if normalized != "" {
				result.WriteString(normalized)
				result.WriteString("\n")
			}

			// Add newline between parts (except for the last part)
			if i < len(parts)-1 && strings.TrimSpace(part) != "" {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

func normalizeWhitespace(line string) string {
	// Replace all tabs with spaces first
	line = strings.ReplaceAll(line, "\t", "    ")

	// Remove excessive leading whitespace but preserve some indentation
	trimmed := strings.TrimLeft(line, " ")
	if trimmed == "" {
		return ""
	}

	// Count leading spaces in the original line
	leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))

	// Normalize to 2-space indentation levels (max 6 levels = 12 spaces)
	indentLevel := leadingSpaces / 2
	if indentLevel > 6 {
		indentLevel = 6 // Cap at 6 levels to prevent excessive indentation
	}

	// Build the normalized line
	var result strings.Builder
	if indentLevel > 0 {
		result.WriteString(strings.Repeat("  ", indentLevel))
	}
	result.WriteString(trimmed)

	return result.String()
}
