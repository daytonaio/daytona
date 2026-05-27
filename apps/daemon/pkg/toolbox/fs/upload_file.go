// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

// UploadFile godoc
//
//	@Summary		Upload a file
//	@Description	Upload a file to the specified path
//	@Tags			file-system
//	@Accept			multipart/form-data
//	@Param			path	query		string	true	"Destination path for the uploaded file"
//	@Param			file	formData	file	true	"File to upload"
//	@Success		200		{object}	gin.H
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/files/upload [post]
//
//	@id				UploadFile
func UploadFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	c.Status(http.StatusOK)
}
