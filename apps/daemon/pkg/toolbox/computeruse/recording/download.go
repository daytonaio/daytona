// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"os"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/recording"
	"github.com/gin-gonic/gin"
)

// DownloadRecording godoc
//
//	@Summary		Download a recording
//	@Description	Download a recording by providing its ID
//	@Tags			computer-use
//	@Produce		octet-stream
//	@Param			id	path		string	true	"Recording ID"
//	@Success		200	{file}		binary
//	@Failure		400	{object}	common.ErrorResponse
//	@Failure		404	{object}	common.ErrorResponse
//	@Failure		500	{object}	common.ErrorResponse
//	@Router			/computeruse/recordings/{id}/download [get]
//
//	@id				DownloadRecording
func (r *RecordingController) DownloadRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("id is required")))
		return
	}

	rec, err := r.recordingService.GetRecording(id)
	if err != nil {
		if errors.Is(err, recording.ErrRecordingNotFound) {
			ctx.Error(common_errors.NewNotFoundError(errors.New("recording not found")))
			return
		}
		ctx.Error(common_errors.NewInternalServerError(err))
		return
	}

	if _, err := os.Stat(rec.FilePath); os.IsNotExist(err) {
		ctx.Error(common_errors.NewNotFoundError(errors.New("recording file not found")))
		return
	} else if err != nil {
		ctx.Error(common_errors.NewInternalServerError(err))
		return
	}

	ctx.File(rec.FilePath)
}
