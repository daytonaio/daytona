// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
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
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Failure		503		{object}	common.ErrorResponse
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
			ctx.Error(common.NewRecordingFfmpegNotFoundError("FFmpeg must be installed and available in PATH to use screen recording"))
			return
		}
		if errors.Is(err, recordingservice.ErrInvalidLabel) {
			ctx.Error(common_errors.NewBadRequestError(err))
			return
		}
		ctx.Error(common_errors.NewInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusCreated, *RecordingToDTO(recording))
}
