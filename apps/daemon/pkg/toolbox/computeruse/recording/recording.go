// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	recordingservice "github.com/daytonaio/daemon/pkg/recording"
)

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
func (r *RecordingController) ListRecordings(ctx *gin.Context) {
	recordings, err := r.recordingService.ListRecordings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recordingDTOs := make([]RecordingDTO, 0, len(recordings))
	for _, rec := range recordings {
		recordingDTOs = append(recordingDTOs, *RecordingToDTO(&rec))
	}

	response := ListRecordingsResponse{
		Recordings: recordingDTOs,
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
//	@Success		200	{object}	RecordingDTO
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id} [get]
//
//	@id				GetRecording
func (r *RecordingController) GetRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := r.recordingService.GetRecording(id)
	if err != nil {
		if errors.Is(err, recordingservice.ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, *RecordingToDTO(recording))
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
func (r *RecordingController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := r.recordingService.DeleteRecording(id)
	if err != nil {
		if errors.Is(err, recordingservice.ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		if errors.Is(err, recordingservice.ErrRecordingStillActive) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete an active recording, stop it first"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
