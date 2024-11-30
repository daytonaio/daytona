// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func DownloadFile(c *gin.Context) {
	requestedPath := c.Query("path")
	if requestedPath == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid path"))
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(404, errors.New("file not found"))
			return
		}
		c.AbortWithError(400, errors.New("unable to access file"))
		return
	}

	if fileInfo.IsDir() {
		c.AbortWithError(400, errors.New("path must be a file"))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(absPath))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.File(absPath)
}
