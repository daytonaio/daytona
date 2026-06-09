// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

// DownloadFile godoc
//
//	@Summary		Download a file
//	@Description	Download a file by providing its path
//	@Tags			file-system
//	@Produce		octet-stream
//	@Param			path	query		string	true	"File path to download"
//	@Success		200		{file}		binary
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		403		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/files/download [get]
//
//	@id				DownloadFile
func DownloadFile(c *gin.Context) {
	requestedPath := c.Query("path")
	if requestedPath == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		c.Error(common_errors.NewBadRequestError(fmt.Errorf("invalid path: %w", err)))
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(common.NewFileNotFoundError(err.Error()))
			return
		}
		if os.IsPermission(err) {
			c.Error(common.NewFileAccessDeniedError(err.Error()))
			return
		}
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	if fileInfo.IsDir() {
		c.Error(common_errors.NewBadRequestError(errors.New("path must be a file")))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	filename := filepath.Base(absPath)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=utf-8''%s`,
		toLatin1(filename), encodeRFC5987(filename)))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.File(absPath)
}
