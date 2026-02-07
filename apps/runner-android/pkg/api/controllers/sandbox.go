// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner-android/internal"
	"github.com/daytonaio/runner-android/pkg/api/dto"
	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/daytonaio/runner-android/pkg/runner"
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
//
//	@Summary		Create a new sandbox
//	@Description	Create a new sandbox from a snapshot
//	@Tags			sandboxes
//	@Accept			json
//	@Produce		json
//	@Param			sandbox	body		dto.CreateSandboxDTO	true	"Sandbox configuration"
//	@Success		201		{object}	dto.StartSandboxResponse
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/sandboxes [post]
func Create(ctx *gin.Context) {
	var createDTO dto.CreateSandboxDTO
	if err := ctx.ShouldBindJSON(&createDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Creating sandbox %s from snapshot %s", createDTO.Id, createDTO.Snapshot)

	// Create the sandbox using Cuttlefish
	_, daemonVersion, err := Runner.CVDClient.Create(ctx.Request.Context(), createDTO)
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
//
//	@Summary		Get sandbox info
//	@Description	Get information about a sandbox
//	@Tags			sandboxes
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{object}	SandboxInfoResponse
//	@Failure		404			{object}	error
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId} [get]
func Info(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	info, err := Runner.CVDClient.GetSandboxInfo(ctx.Request.Context(), sandboxId)
	if err != nil {
		log.Errorf("Failed to get sandbox info for %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Map CVD state to sandbox state
	state := mapCVDStateToSandboxState(info.State)

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

// mapCVDStateToSandboxState maps Cuttlefish state to sandbox state
func mapCVDStateToSandboxState(cvdState cuttlefish.InstanceState) SandboxState {
	switch cvdState {
	case cuttlefish.InstanceStateRunning:
		return SandboxStateStarted
	case cuttlefish.InstanceStateStopped:
		return SandboxStateStopped
	case cuttlefish.InstanceStateStarting:
		return SandboxStateCreating
	default:
		return SandboxStateUnknown
	}
}

// Destroy destroys a sandbox
//
//	@Summary		Destroy a sandbox
//	@Description	Destroy a sandbox and all its resources
//	@Tags			sandboxes
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox destroyed"
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/destroy [post]
func Destroy(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Destroying sandbox %s", sandboxId)

	if err := Runner.CVDClient.Destroy(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to destroy sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox destroyed")
}

// Start starts a sandbox
//
//	@Summary		Start a sandbox
//	@Description	Start a stopped or paused sandbox
//	@Tags			sandboxes
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{object}	dto.StartSandboxResponse
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/start [post]
func Start(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Starting sandbox %s", sandboxId)

	daemonVersion, err := Runner.CVDClient.Start(ctx.Request.Context(), sandboxId, nil)
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
//
//	@Summary		Stop a sandbox
//	@Description	Stop (pause) a running sandbox
//	@Tags			sandboxes
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox stopped"
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/stop [post]
func Stop(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Stopping sandbox %s", sandboxId)

	if err := Runner.CVDClient.Stop(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to stop sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox stopped")
}

// Resize resizes a sandbox's resources
//
//	@Summary		Resize a sandbox
//	@Description	Resize CPU, memory, or GPU allocation for a sandbox (not supported for Cuttlefish)
//	@Tags			sandboxes
//	@Accept			json
//	@Param			sandboxId	path		string					true	"Sandbox ID"
//	@Param			resize		body		dto.ResizeSandboxDTO	true	"New resource allocation"
//	@Success		200			{string}	string					"Sandbox resized"
//	@Failure		400			{object}	error
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/resize [post]
func Resize(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var resizeDTO dto.ResizeSandboxDTO
	if err := ctx.ShouldBindJSON(&resizeDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Resize not supported for Cuttlefish sandbox %s", sandboxId)

	// Cuttlefish doesn't support live resizing
	ctx.JSON(http.StatusOK, "Resize not supported for Cuttlefish")
}

// RemoveDestroyed removes a destroyed sandbox's resources
//
//	@Summary		Remove destroyed sandbox
//	@Description	Clean up resources for a destroyed sandbox
//	@Tags			sandboxes
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox removed"
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId} [delete]
func RemoveDestroyed(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Removing destroyed sandbox %s", sandboxId)

	if err := Runner.CVDClient.RemoveDestroyed(ctx.Request.Context(), sandboxId); err != nil {
		log.Errorf("Failed to remove destroyed sandbox %s: %v", sandboxId, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox removed")
}

// UpdateNetworkSettings updates network settings for a sandbox
//
//	@Summary		Update network settings
//	@Description	Update network blocking and allow list settings
//	@Tags			sandboxes
//	@Accept			json
//	@Param			sandboxId	path		string							true	"Sandbox ID"
//	@Param			settings	body		dto.UpdateNetworkSettingsDTO	true	"Network settings"
//	@Success		200			{string}	string							"Network settings updated"
//	@Failure		400			{object}	error
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/network-settings [post]
func UpdateNetworkSettings(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var settingsDTO dto.UpdateNetworkSettingsDTO
	if err := ctx.ShouldBindJSON(&settingsDTO); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("Updating network settings for sandbox %s (no-op for Cuttlefish)", sandboxId)

	// Cuttlefish manages its own networking
	ctx.JSON(http.StatusOK, "Network settings updated")
}

// CreateBackup creates a backup/snapshot of a sandbox
//
//	@Summary		Create backup
//	@Description	Create a backup (snapshot) of a running sandbox
//	@Tags			sandboxes
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		201			{string}	string	"Backup started"
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/backup [post]
func CreateBackup(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("CreateBackup not supported for Cuttlefish sandbox %s", sandboxId)

	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Backup not supported for Cuttlefish"})
}

// Fork creates a fork (copy-on-write clone) of a running sandbox
//
//	@Summary		Fork a sandbox
//	@Description	Create a CoW fork of a running/paused sandbox with memory state (not supported for Cuttlefish)
//	@Tags			sandboxes
//	@Accept			json
//	@Produce		json
//	@Param			sandboxId	path		string				true	"Source Sandbox ID"
//	@Param			fork		body		dto.ForkSandboxDTO	true	"Fork configuration"
//	@Success		201			{object}	dto.ForkSandboxResponseDTO
//	@Failure		400			{object}	error
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/fork [post]
func Fork(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Fork not supported for Cuttlefish sandbox %s", sandboxId)

	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Fork not supported for Cuttlefish"})
}

// Clone creates a complete copy of a sandbox with flattened filesystem
//
//	@Summary		Clone a sandbox
//	@Description	Create an independent copy of a sandbox (not supported for Cuttlefish)
//	@Tags			sandboxes
//	@Accept			json
//	@Produce		json
//	@Param			sandboxId	path		string					true	"Source Sandbox ID"
//	@Param			clone		body		dto.CloneSandboxDTO		true	"Clone configuration"
//	@Success		201			{object}	dto.CloneSandboxResponseDTO
//	@Failure		400			{object}	error
//	@Failure		500			{object}	error
//	@Router			/sandboxes/{sandboxId}/clone [post]
func Clone(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	log.Infof("Clone not supported for Cuttlefish sandbox %s", sandboxId)

	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Clone not supported for Cuttlefish"})
}

// HealthCheck returns health status
//
//	@Summary		Health check
//	@Description	Returns OK if the runner is healthy
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/ [get]
func HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": internal.Version,
	})
}

// RunnerInfo returns runner information
//
//	@Summary		Runner info
//	@Description	Returns information about the runner
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/info [get]
func RunnerInfo(ctx *gin.Context) {
	// Get list of sandboxes
	sandboxes, _ := Runner.CVDClient.List(ctx.Request.Context())

	ctx.JSON(http.StatusOK, gin.H{
		"type":      "cuttlefish",
		"sandboxes": len(sandboxes),
	})
}

// NOTE: Memory stats endpoints moved to stats.go

// NOTE: ProxyRequest and ProxyToPort are implemented in proxy.go

// PullSnapshot pulls a snapshot from registry
//
//	@Summary		Pull snapshot
//	@Description	Pulls a snapshot from the registry
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.PullSnapshotRequestDTO	true	"Pull request"
//	@Success		200
//	@Router			/snapshots/pull [post]
func PullSnapshot(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot pull not supported for Cuttlefish"})
}

// PushSnapshot pushes a snapshot to registry
//
//	@Summary		Push snapshot
//	@Description	Pushes a snapshot to the registry
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.PushSnapshotRequestDTO	true	"Push request"
//	@Success		200
//	@Router			/snapshots/push [post]
func PushSnapshot(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot push not supported for Cuttlefish"})
}

// CreateSnapshot creates a new snapshot from a sandbox
//
//	@Summary		Create snapshot
//	@Description	Creates a snapshot from a running sandbox
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.CreateSnapshotRequestDTO	true	"Create request"
//	@Success		200
//	@Router			/snapshots/create [post]
func CreateSnapshot(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot creation not supported for Cuttlefish"})
}

// BuildSnapshot builds a snapshot from a Dockerfile
//
//	@Summary		Build snapshot
//	@Description	Builds a snapshot from a Dockerfile
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.BuildSnapshotRequestDTO	true	"Build request"
//	@Success		200
//	@Router			/snapshots/build [post]
func BuildSnapshot(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot build not supported for Cuttlefish"})
}

// TagImage tags an image
//
//	@Summary		Tag image
//	@Description	Tags an image with a new reference
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.TagImageRequestDTO	true	"Tag request"
//	@Success		200
//	@Router			/snapshots/tag [post]
func TagImage(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Image tagging not supported for Cuttlefish"})
}

// SnapshotExists checks if a snapshot exists
//
//	@Summary		Check snapshot exists
//	@Description	Checks if a snapshot exists
//	@Tags			snapshots
//	@Param			ref	query		string	true	"Snapshot reference"
//	@Success		200	{object}	bool
//	@Router			/snapshots/exists [get]
func SnapshotExists(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
		return
	}

	exists, err := Runner.CVDClient.SnapshotExists(ctx.Request.Context(), name)
	if err != nil {
		log.Errorf("Failed to check snapshot %s: %v", name, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, exists)
}

// GetSnapshotInfo returns snapshot information
//
//	@Summary		Get snapshot info
//	@Description	Returns information about a snapshot
//	@Tags			snapshots
//	@Param			ref	query		string	true	"Snapshot reference"
//	@Success		200	{object}	dto.SnapshotInfoResponseDTO
//	@Router			/snapshots/info [get]
func GetSnapshotInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot info not supported for Cuttlefish"})
}

// RemoveSnapshot removes a snapshot
//
//	@Summary		Remove snapshot
//	@Description	Removes a snapshot
//	@Tags			snapshots
//	@Accept			json
//	@Param			request	body	dto.RemoveImageRequestDTO	true	"Remove request"
//	@Success		200
//	@Router			/snapshots/remove [post]
func RemoveSnapshot(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Snapshot removal not supported for Cuttlefish"})
}

// GetBuildLogs returns build logs
//
//	@Summary		Get build logs
//	@Description	Returns build logs for a snapshot build
//	@Tags			snapshots
//	@Success		200	{string}	string
//	@Router			/snapshots/logs [get]
func GetBuildLogs(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Build logs not supported for Cuttlefish"})
}
