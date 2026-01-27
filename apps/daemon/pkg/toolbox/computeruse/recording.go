// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Recording represents a recording session (active or completed)
type Recording struct {
	ID              string     `json:"id"`
	FileName        string     `json:"fileName"`
	FilePath        string     `json:"filePath"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	Status          string     `json:"status"`
	DurationSeconds *float64   `json:"durationSeconds,omitempty"`
	SizeBytes       *int64     `json:"sizeBytes,omitempty"`
} // @name Recording

// StartRecordingRequest represents the request to start a new recording
type StartRecordingRequest struct {
	Label string `json:"label,omitempty"` // Optional custom label for the recording
} // @name StartRecordingRequest

// StopRecordingRequest represents the request to stop an active recording
type StopRecordingRequest struct {
	ID string `json:"id" validate:"required"` // Recording ID to stop
} // @name StopRecordingRequest

// StartRecordingResponse represents the response after starting a recording
type StartRecordingResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	StartTime time.Time `json:"startTime"`
	Status    string    `json:"status"`
} // @name StartRecordingResponse

// StopRecordingResponse represents the response after stopping a recording
type StopRecordingResponse struct {
	ID              string  `json:"id"`
	FilePath        string  `json:"filePath"`
	DurationSeconds float64 `json:"durationSeconds"`
	Status          string  `json:"status"`
} // @name StopRecordingResponse

// ListRecordingsResponse represents the response containing all recordings
type ListRecordingsResponse struct {
	Recordings []Recording `json:"recordings"`
} // @name ListRecordingsResponse

// GetRecordingResponse represents the response for a single recording
type GetRecordingResponse struct {
	Recording
} // @name GetRecordingResponse

// RecordingHandler handles recording-related HTTP requests
type RecordingHandler struct {
	manager *RecordingManager
}

// NewRecordingHandler creates a new RecordingHandler
func NewRecordingHandler() *RecordingHandler {
	return &RecordingHandler{
		manager: GetRecordingManager(),
	}
}

// StartRecording godoc
//
//	@Summary		Start a new recording
//	@Description	Start a new screen recording session
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StartRecordingRequest	false	"Recording options"
//	@Success		201		{object}	StartRecordingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/computeruse/recordings/start [post]
//
//	@id				StartRecording
func (h *RecordingHandler) StartRecording(ctx *gin.Context) {
	var request StartRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Allow empty body - label is optional
		request = StartRecordingRequest{}
	}

	recording, err := h.manager.StartRecording(request.Label)
	if err != nil {
		if errors.Is(err, ErrFFmpegNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "ffmpeg_not_found",
				"message": "FFmpeg must be installed and available in PATH to use screen recording",
			})
			return
		}
		if errors.Is(err, ErrNoDisplay) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "no_display",
				"message": "DISPLAY environment variable not set - X11 display required for screen recording",
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
		Status:    recording.Status,
	}

	ctx.JSON(http.StatusCreated, response)
}

// StopRecording godoc
//
//	@Summary		Stop a recording
//	@Description	Stop an active screen recording session
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StopRecordingRequest	true	"Recording ID to stop"
//	@Success		200		{object}	StopRecordingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/computeruse/recordings/stop [post]
//
//	@id				StopRecording
func (h *RecordingHandler) StopRecording(ctx *gin.Context) {
	var request StopRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: id is required"})
		return
	}

	if request.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := h.manager.StopRecording(request.ID)
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
		Status:          recording.Status,
	}

	ctx.JSON(http.StatusOK, response)
}

// ListRecordings godoc
//
//	@Summary		List all recordings
//	@Description	Get a list of all recordings (active and completed)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ListRecordingsResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings [get]
//
//	@id				ListRecordings
func (h *RecordingHandler) ListRecordings(ctx *gin.Context) {
	recordings, err := h.manager.ListRecordings()
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
//	@Tags			computer-use
//	@Produce		json
//	@Param			id	path		string	true	"Recording ID"
//	@Success		200	{object}	GetRecordingResponse
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [get]
//
//	@id				GetRecording
func (h *RecordingHandler) GetRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := h.manager.GetRecording(id)
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
//	@Tags			computer-use
//	@Param			id	path	string	true	"Recording ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [delete]
//
//	@id				DeleteRecording
func (h *RecordingHandler) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := h.manager.DeleteRecording(id)
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
