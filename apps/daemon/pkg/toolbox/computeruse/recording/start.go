// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	recordingservice "github.com/daytonaio/daemon/pkg/recording"
)

// StartRecording godoc
//
//	@Summary		Start a new recording
//	@Description	Start a new screen recording session
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StartRecordingRequest	false	"Recording options"
//	@Success		201		{object}	RecordingDTO
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/computeruse/recordings/start [post]
//
//	@id				StartRecording
func (h *RecordingController) StartRecording(ctx *gin.Context) {
	var request StartRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Allow empty body - label is optional
		request = StartRecordingRequest{}
	}

	recording, err := h.recordingService.StartRecording(request.Label)
	if err != nil {
		if errors.Is(err, recordingservice.ErrFFmpegNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "ffmpeg_not_found",
				"message": "FFmpeg must be installed and available in PATH to use screen recording",
			})
			return
		}
		if errors.Is(err, recordingservice.ErrInvalidLabel) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_label",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, *RecordingToDTO(recording))
}
