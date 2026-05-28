// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
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
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/computeruse/recordings/stop [post]
//
//	@id				StopRecording
func (r *RecordingController) StopRecording(ctx *gin.Context) {
	var request StopRecordingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	if request.ID == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("id is required")))
		return
	}

	recording, err := r.recordingService.StopRecording(request.ID)
	if err != nil {
		if errors.Is(err, recordingservice.ErrRecordingNotFound) {
			ctx.Error(common_errors.NewNotFoundError(errors.New("recording not found")))
			return
		}
		ctx.Error(common_errors.NewInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusOK, *RecordingToDTO(recording))
}
