// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"

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
//	@Router			/files/upload [post]
//
//	@id				UploadFile
func UploadFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusOK)
}
