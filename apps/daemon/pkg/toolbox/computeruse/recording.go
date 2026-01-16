// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
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
type RecordingHandler struct{}

// NewRecordingHandler creates a new RecordingHandler
func NewRecordingHandler() *RecordingHandler {
	return &RecordingHandler{}
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
//	@Failure		501		{object}	map[string]string
//	@Router			/computeruse/recordings/start [post]
//
//	@id				StartRecording
func (h *RecordingHandler) StartRecording(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Screen recording is not implemented on this platform",
	})
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
//	@Failure		501		{object}	map[string]string
//	@Router			/computeruse/recordings/stop [post]
//
//	@id				StopRecording
func (h *RecordingHandler) StopRecording(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Screen recording is not implemented on this platform",
	})
}

// ListRecordings godoc
//
//	@Summary		List all recordings
//	@Description	Get a list of all recordings (active and completed)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ListRecordingsResponse
//	@Failure		501	{object}	map[string]string
//	@Router			/computeruse/recordings [get]
//
//	@id				ListRecordings
func (h *RecordingHandler) ListRecordings(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Screen recording is not implemented on this platform",
	})
}

// GetRecording godoc
//
//	@Summary		Get recording details
//	@Description	Get details of a specific recording by ID
//	@Tags			computer-use
//	@Produce		json
//	@Param			id	path		string	true	"Recording ID"
//	@Success		200	{object}	GetRecordingResponse
//	@Failure		501	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [get]
//
//	@id				GetRecording
func (h *RecordingHandler) GetRecording(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Screen recording is not implemented on this platform",
	})
}

// DeleteRecording godoc
//
//	@Summary		Delete a recording
//	@Description	Delete a recording file by ID
//	@Tags			computer-use
//	@Param			id	path	string	true	"Recording ID"
//	@Success		204
//	@Failure		501	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [delete]
//
//	@id				DeleteRecording
func (h *RecordingHandler) DeleteRecording(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Screen recording is not implemented on this platform",
	})
}
