// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecordingController handles recording-related HTTP requests
type RecordingController struct {
	manager *RecordingManager
}

// NewRecordingController creates a new RecordingController
func NewRecordingController(configDir string) *RecordingController {
	return &RecordingController{
		manager: NewRecordingManager(configDir),
	}
}

// StartRecording godoc
//
//	@Summary		Start a new recording
//	@Description	Start a new screen recording session
//	@Tags			recording
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StartRecordingRequest	false	"Recording options"
//	@Success		201		{object}	StartRecordingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/computeruse/recordings/start [post]
//
//	@id				StartRecording
func (c *RecordingController) StartRecording(ctx *gin.Context) {
	var request StartRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Allow empty body - label is optional
		request = StartRecordingRequest{}
	}

	recording, err := c.manager.StartRecording(request.Label)
	if err != nil {
		if errors.Is(err, ErrFFmpegNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "ffmpeg not found",
				"message": "FFmpeg must be installed and available in PATH to use screen recording",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := StartRecordingResponse{
		ID:        recording.ID,
		FileName:  recording.FileName,
		FilePath:  recording.FilePath,
		StartTime: recording.StartTime,
		Status:    string(recording.Status),
	}

	ctx.JSON(http.StatusCreated, response)
}

// StopRecording godoc
//
//	@Summary		Stop a recording
//	@Description	Stop an active screen recording session
//	@Tags			recording
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StopRecordingRequest	true	"Recording ID to stop"
//	@Success		200		{object}	StopRecordingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/computeruse/recordings/stop [post]
//
//	@id				StopRecording
func (c *RecordingController) StopRecording(ctx *gin.Context) {
	var request StopRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: id is required"})
		return
	}

	if request.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := c.manager.StopRecording(request.ID)
	if err != nil {
		if errors.Is(err, ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	durationSeconds := float64(0)
	if recording.DurationSeconds != nil {
		durationSeconds = *recording.DurationSeconds
	}

	response := StopRecordingResponse{
		ID:              recording.ID,
		FilePath:        recording.FilePath,
		DurationSeconds: durationSeconds,
		Status:          string(recording.Status),
	}

	ctx.JSON(http.StatusOK, response)
}

// ListRecordings godoc
//
//	@Summary		List all recordings
//	@Description	Get a list of all recordings (active and completed)
//	@Tags			recording
//	@Produce		json
//	@Success		200	{object}	ListRecordingsResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings [get]
//
//	@id				ListRecordings
func (c *RecordingController) ListRecordings(ctx *gin.Context) {
	recordings, err := c.manager.ListRecordings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ListRecordingsResponse{
		Recordings: recordings,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetRecording godoc
//
//	@Summary		Get recording details
//	@Description	Get details of a specific recording by ID
//	@Tags			recording
//	@Produce		json
//	@Param			id	path		string	true	"Recording ID"
//	@Success		200	{object}	GetRecordingResponse
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [get]
//
//	@id				GetRecording
func (c *RecordingController) GetRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := c.manager.GetRecording(id)
	if err != nil {
		if errors.Is(err, ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := GetRecordingResponse{
		Recording: *recording,
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteRecording godoc
//
//	@Summary		Delete a recording
//	@Description	Delete a recording file by ID
//	@Tags			recording
//	@Param			id	path	string	true	"Recording ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [delete]
//
//	@id				DeleteRecording
func (c *RecordingController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := c.manager.DeleteRecording(id)
	if err != nil {
		if errors.Is(err, ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		if errors.Is(err, ErrRecordingStillActive) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete an active recording, stop it first"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
