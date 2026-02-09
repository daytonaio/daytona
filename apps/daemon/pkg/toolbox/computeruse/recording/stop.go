// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	recordingservice "github.com/daytonaio/daemon/pkg/recording"
)

// StopRecording godoc
//
//	@Summary		Stop a recording
//	@Description	Stop an active screen recording session
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StopRecordingRequest	true	"Recording ID to stop"
//	@Success		200		{object}	RecordingDTO
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/computeruse/recordings/stop [post]
//
//	@id				StopRecording
func (r *RecordingController) StopRecording(ctx *gin.Context) {
	var request StopRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: id is required"})
		return
	}

	if request.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	recording, err := r.recordingService.StopRecording(request.ID)
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
