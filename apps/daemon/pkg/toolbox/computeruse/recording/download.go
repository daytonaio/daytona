// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"net/http"
	"os"

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
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/computeruse/recordings/{id}/download [get]
//
//	@id				DownloadRecording
func (r *RecordingController) DownloadRecording(ctx *gin.Context) {
	id := ctx.Param("id")

	rec, err := r.recordingService.GetRecording(id)
	if err != nil {
		if errors.Is(err, recording.ErrRecordingNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(rec.FilePath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	ctx.File(rec.FilePath)
}
