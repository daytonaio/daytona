// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner-ch/internal"
	"github.com/daytonaio/runner-ch/pkg/api/dto"
	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/runner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var Runner *runner.Runner

// SandboxState represents the state of a sandbox
type SandboxState string

const (
	SandboxStateStarted  SandboxState = "started"
	SandboxStateStopped  SandboxState = "stopped"
	SandboxStateError    SandboxState = "error"
	SandboxStateCreating SandboxState = "creating"
	SandboxStateUnknown  SandboxState = "unknown"
)

// BackupState represents the state of a backup
type BackupState string

const (
	BackupStateNone       BackupState = "none"
	BackupStateInProgress BackupState = "in_progress"
	BackupStateCompleted  BackupState = "completed"
	BackupStateFailed     BackupState = "failed"
)

// SandboxInfoResponse matches runner-win's response format
type SandboxInfoResponse struct {
	State         SandboxState `json:"state"`
	BackupState   BackupState  `json:"backupState"`
	BackupError   *string      `json:"backupError,omitempty"`
	DaemonVersion *string      `json:"daemonVersion,omitempty"`
} //	@name	SandboxInfoResponse

// Create creates a new sandbox
// @Summary		Create a new sandbox
// @Description	Create a new sandbox from a snapshot
// @Tags			sandboxes
// @Accept			json
// @Produce		json
// @Param			sandbox	body		dto.CreateSandboxDTO	true	"Sandbox configuration"
// @Success		201		{object}	dto.StartSandboxResponse
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Router			/sandboxes [post]
func Create(ctx *gin.Context) {
	var createDTO dto.CreateSandboxDTO
	if err := ctx.ShouldBindJSON(&createDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Creating sandbox %s from snapshot %s", createDTO.Id, createDTO.Snapshot)

	// Create the sandbox using Cloud Hypervisor
	_, daemonVersion, err := Runner.CHClient.Create(ctx.Request.Context(), createDTO)
	if err != nil {
		log.Errorf("Failed to create sandbox %s: %v", createDTO.Id, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	})
}

// Info returns information about a sandbox
// @Summary		Get sandbox info
// @Description	Get information about a sandbox
// @Tags			sandboxes
// @Produce		json
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		200			{object}	SandboxInfoResponse
// @Failure		404			{object}	error
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId} [get]
func Info(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	info, err := Runner.CHClient.GetSandboxInfo(ctx.Request.Context(), sandboxId)
	if err != nil {
		log.Errorf("Failed to get sandbox info for %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Map CH state to sandbox state
	state := mapCHStateToSandboxState(info.State)

	var daemonVersion *string
	if state == SandboxStateStarted {
		v := internal.Version
		daemonVersion = &v
	}

	ctx.JSON(http.StatusOK, SandboxInfoResponse{
		State:         state,
		BackupState:   BackupStateNone,
		BackupError:   nil,
		DaemonVersion: daemonVersion,
	})
}

// mapCHStateToSandboxState maps Cloud Hypervisor state to sandbox state
func mapCHStateToSandboxState(chState cloudhypervisor.VmState) SandboxState {
	switch chState {
	case cloudhypervisor.VmStateRunning:
		return SandboxStateStarted
	case cloudhypervisor.VmStatePaused, cloudhypervisor.VmStateShutdown, cloudhypervisor.VmStateCreated:
		return SandboxStateStopped
	default:
		return SandboxStateUnknown
	}
}

// Destroy destroys a sandbox
// @Summary		Destroy a sandbox
// @Description	Destroy a sandbox and all its resources
// @Tags			sandboxes
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		200			{string}	string	"Sandbox destroyed"
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/destroy [post]
func Destroy(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Destroying sandbox %s", sandboxId)

	if err := Runner.CHClient.Destroy(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to destroy sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox destroyed")
}

// Start starts a sandbox
// @Summary		Start a sandbox
// @Description	Start a stopped or paused sandbox
// @Tags			sandboxes
// @Produce		json
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		200			{object}	dto.StartSandboxResponse
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/start [post]
func Start(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Starting sandbox %s", sandboxId)

	daemonVersion, err := Runner.CHClient.Start(ctx.Request.Context(), sandboxId, nil)
	if err != nil {
		log.Errorf("Failed to start sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.StartSandboxResponse{
		DaemonVersion: daemonVersion,
	})
}

// Stop stops a sandbox
// @Summary		Stop a sandbox
// @Description	Stop (pause) a running sandbox
// @Tags			sandboxes
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		200			{string}	string	"Sandbox stopped"
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/stop [post]
func Stop(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Stopping sandbox %s", sandboxId)

	if err := Runner.CHClient.Stop(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to stop sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox stopped")
}

// Resize resizes a sandbox's resources
// @Summary		Resize a sandbox
// @Description	Resize CPU, memory, or GPU allocation for a sandbox
// @Tags			sandboxes
// @Accept			json
// @Param			sandboxId	path		string				true	"Sandbox ID"
// @Param			resize		body		dto.ResizeSandboxDTO	true	"New resource allocation"
// @Success		200			{string}	string	"Sandbox resized"
// @Failure		400			{object}	error
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/resize [post]
func Resize(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var resizeDTO dto.ResizeSandboxDTO
	if err := ctx.ShouldBindJSON(&resizeDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Resizing sandbox %s: cpu=%d, memory=%d, gpu=%d",
		sandboxId, resizeDTO.Cpu, resizeDTO.Memory, resizeDTO.Gpu)

	// Cloud Hypervisor supports live resizing
	if err := Runner.CHClient.Resize(ctx.Request.Context(), sandboxId,
		int(resizeDTO.Cpu), uint64(resizeDTO.Memory)); err != nil {
		log.Errorf("Failed to resize sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox resized")
}

// RemoveDestroyed removes a destroyed sandbox's resources
// @Summary		Remove destroyed sandbox
// @Description	Clean up resources for a destroyed sandbox
// @Tags			sandboxes
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		200			{string}	string	"Sandbox removed"
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId} [delete]
func RemoveDestroyed(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Removing destroyed sandbox %s", sandboxId)

	if err := Runner.CHClient.RemoveDestroyed(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to remove destroyed sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox removed")
}

// UpdateNetworkSettings updates network settings for a sandbox
// @Summary		Update network settings
// @Description	Update network blocking and allow list settings
// @Tags			sandboxes
// @Accept			json
// @Param			sandboxId	path		string						true	"Sandbox ID"
// @Param			settings	body		dto.UpdateNetworkSettingsDTO	true	"Network settings"
// @Success		200			{string}	string	"Network settings updated"
// @Failure		400			{object}	error
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/network-settings [post]
func UpdateNetworkSettings(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var settingsDTO dto.UpdateNetworkSettingsDTO
	if err := ctx.ShouldBindJSON(&settingsDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Updating network settings for sandbox %s", sandboxId)

	// Get sandbox IP
	info, err := Runner.CHClient.GetSandboxInfo(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Update network rules
	if err := Runner.NetRulesManager.UpdateNetworkSettings(sandboxId, info.IpAddress,
		settingsDTO.NetworkBlockAll, settingsDTO.NetworkAllowList); err != nil {
		log.Errorf("Failed to update network settings for %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Network settings updated")
}

// CreateBackup creates a backup/snapshot of a sandbox
// @Summary		Create backup
// @Description	Create a backup (snapshot) of a running sandbox
// @Tags			sandboxes
// @Param			sandboxId	path		string	true	"Sandbox ID"
// @Success		201			{string}	string	"Backup started"
// @Failure		500			{object}	error
// @Router			/sandboxes/{sandboxId}/backup [post]
func CreateBackup(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Creating backup for sandbox %s", sandboxId)

	_, err := Runner.CHClient.CreateSnapshot(ctx.Request.Context(), dto.CreateSnapshotRequestDTO{
		SandboxId: sandboxId,
		Name:      sandboxId + "-backup",
	})
	if err != nil {
		log.Errorf("Failed to create backup for %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, "Backup started")
}

// HealthCheck returns health status
// @Summary		Health check
// @Description	Returns OK if the runner is healthy
// @Tags			health
// @Produce		json
// @Success		200	{object}	map[string]string
// @Router			/ [get]
func HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": internal.Version,
	})
}

// RunnerInfo returns runner information
// @Summary		Runner info
// @Description	Returns information about the runner
// @Tags			info
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Router			/info [get]
func RunnerInfo(ctx *gin.Context) {
	// Get list of sandboxes
	sandboxes, _ := Runner.CHClient.List(ctx.Request.Context())

	ctx.JSON(http.StatusOK, gin.H{
		"type":      "cloud-hypervisor",
		"sandboxes": len(sandboxes),
	})
}

// GetMemoryStats returns memory statistics
// @Summary		Get memory stats
// @Description	Returns memory usage statistics
// @Tags			stats
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Router			/stats/memory [get]
func GetMemoryStats(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"total":     0,
		"used":      0,
		"free":      0,
		"available": 0,
	})
}

// GetMemoryStatsView returns memory statistics view
// @Summary		Get memory stats view
// @Description	Returns memory usage statistics in HTML view
// @Tags			stats
// @Produce		html
// @Success		200	{string}	string
// @Router			/stats/memory/view [get]
func GetMemoryStatsView(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Memory stats view not implemented for Cloud Hypervisor runner")
}

// ProxyRequest proxies requests to the sandbox toolbox
// @Summary		Proxy to toolbox
// @Description	Proxies requests to the sandbox's toolbox daemon
// @Tags			sandboxes
// @Param			sandboxId	path	string	true	"Sandbox ID"
// @Param			path		path	string	true	"Path"
// @Router			/sandboxes/{sandboxId}/toolbox/{path} [get]
func ProxyRequest(ctx *gin.Context) {
	// TODO: Implement toolbox proxy
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Toolbox proxy not yet implemented"})
}

// ProxyToPort proxies requests to a specific port in the sandbox
// @Summary		Proxy to port
// @Description	Proxies requests to a specific port in the sandbox
// @Tags			sandboxes
// @Param			sandboxId	path	string	true	"Sandbox ID"
// @Param			port		path	int		true	"Port"
// @Param			path		path	string	true	"Path"
// @Router			/sandboxes/{sandboxId}/proxy/{port}/{path} [get]
func ProxyToPort(ctx *gin.Context) {
	// TODO: Implement port proxy
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Port proxy not yet implemented"})
}

// PullSnapshot pulls a snapshot from registry
// @Summary		Pull snapshot
// @Description	Pulls a snapshot from the registry
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.PullSnapshotRequestDTO	true	"Pull request"
// @Success		200
// @Router			/snapshots/pull [post]
func PullSnapshot(ctx *gin.Context) {
	// TODO: Implement snapshot pull
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot pull not yet implemented"})
}

// PushSnapshot pushes a snapshot to registry
// @Summary		Push snapshot
// @Description	Pushes a snapshot to the registry
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.PushSnapshotRequestDTO	true	"Push request"
// @Success		200
// @Router			/snapshots/push [post]
func PushSnapshot(ctx *gin.Context) {
	// TODO: Implement snapshot push
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot push not yet implemented"})
}

// CreateSnapshot creates a new snapshot from a sandbox
// @Summary		Create snapshot
// @Description	Creates a snapshot from a running sandbox
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.CreateSnapshotRequestDTO	true	"Create request"
// @Success		200
// @Router			/snapshots/create [post]
func CreateSnapshot(ctx *gin.Context) {
	var createDTO dto.CreateSnapshotRequestDTO
	if err := ctx.ShouldBindJSON(&createDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := Runner.CHClient.CreateSnapshot(ctx.Request.Context(), createDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BuildSnapshot builds a snapshot from a Dockerfile
// @Summary		Build snapshot
// @Description	Builds a snapshot from a Dockerfile
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.BuildSnapshotRequestDTO	true	"Build request"
// @Success		200
// @Router			/snapshots/build [post]
func BuildSnapshot(ctx *gin.Context) {
	// TODO: Implement snapshot build
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot build not yet implemented"})
}

// TagImage tags an image
// @Summary		Tag image
// @Description	Tags an image with a new reference
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.TagImageRequestDTO	true	"Tag request"
// @Success		200
// @Router			/snapshots/tag [post]
func TagImage(ctx *gin.Context) {
	// TODO: Implement image tagging
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Image tagging not yet implemented"})
}

// SnapshotExists checks if a snapshot exists
// @Summary		Check snapshot exists
// @Description	Checks if a snapshot exists
// @Tags			snapshots
// @Param			ref	query	string	true	"Snapshot reference"
// @Success		200	{object}	bool
// @Router			/snapshots/exists [get]
func SnapshotExists(ctx *gin.Context) {
	ref := ctx.Query("ref")
	if ref == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ref query parameter required"})
		return
	}

	snapshots, err := Runner.CHClient.ListSnapshots(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, s := range snapshots {
		if s == ref {
			ctx.JSON(http.StatusOK, true)
			return
		}
	}

	ctx.JSON(http.StatusOK, false)
}

// GetSnapshotInfo returns snapshot information
// @Summary		Get snapshot info
// @Description	Returns information about a snapshot
// @Tags			snapshots
// @Param			ref	query	string	true	"Snapshot reference"
// @Success		200	{object}	dto.SnapshotInfoResponseDTO
// @Router			/snapshots/info [get]
func GetSnapshotInfo(ctx *gin.Context) {
	ref := ctx.Query("ref")
	if ref == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ref query parameter required"})
		return
	}

	info, err := Runner.CHClient.GetSnapshotInfo(ctx.Request.Context(), ref)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SnapshotInfoResponseDTO{
		Size:    info.DiskSizeBytes,
		Created: info.CreatedAt.String(),
	})
}

// RemoveSnapshot removes a snapshot
// @Summary		Remove snapshot
// @Description	Removes a snapshot
// @Tags			snapshots
// @Accept			json
// @Param			request	body	dto.RemoveImageRequestDTO	true	"Remove request"
// @Success		200
// @Router			/snapshots/remove [post]
func RemoveSnapshot(ctx *gin.Context) {
	var removeDTO dto.RemoveImageRequestDTO
	if err := ctx.ShouldBindJSON(&removeDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := Runner.CHClient.DeleteSnapshot(ctx.Request.Context(), removeDTO.Ref); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}

// GetBuildLogs returns build logs
// @Summary		Get build logs
// @Description	Returns build logs for a snapshot build
// @Tags			snapshots
// @Success		200	{string}	string
// @Router			/snapshots/logs [get]
func GetBuildLogs(ctx *gin.Context) {
	// TODO: Implement build logs
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Build logs not yet implemented"})
}
