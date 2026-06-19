// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type UploadedFile struct {
	Name string `json:"name" validate:"required"`
	Path string `json:"path" validate:"required"`
	Type string `json:"type" validate:"required"`
} //	@name	UploadedFile

// UploadFile is the legacy upload handler. It stays registered at /files/upload
// for backward compatibility with older SDK clients but is no longer documented
// in the OpenAPI spec (no swag annotations). New clients use UploadFileV2.
func UploadFile(c *gin.Context) {
	uploadFile(c, false)
}

// UploadFileV2 godoc
//
//	@Summary		Upload a file
//	@Description	Upload a file to the specified path. Accepts either multipart/form-data
//	@Description	(field "file") or a raw request body (e.g. application/octet-stream).
//	@Description	Parent directories are created if missing; an existing file is overwritten.
//	@Tags			file-system
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			path	query		string	true	"Destination path for the uploaded file"
//	@Param			file	formData	file	false	"File to upload (multipart/form-data)"
//	@Success		200		{object}	UploadedFile
//	@Router			/files/upload-v2 [post]
//
//	@id				UploadFile
func UploadFileV2(c *gin.Context) {
	uploadFile(c, true)
}

func uploadFile(c *gin.Context, returnInfo bool) {
	enableFullDuplex(c)
	defer drainBody(c)

	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	} else {
		dst, err := os.Create(path)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if _, err := io.Copy(dst, c.Request.Body); err != nil {
			dst.Close()
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := dst.Close(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if returnInfo {
		c.JSON(http.StatusOK, UploadedFile{
			Name: filepath.Base(path),
			Path: path,
			Type: "file",
		})
	} else {
		c.Status(http.StatusOK)
	}
}
