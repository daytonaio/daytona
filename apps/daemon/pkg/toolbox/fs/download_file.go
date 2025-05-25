// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func DownloadFile(c *gin.Context) {
	requestedPath := c.Query("path")
	if requestedPath == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid path: %w", err))
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if os.IsPermission(err) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if fileInfo.IsDir() {
		c.AbortWithError(http.StatusBadRequest, errors.New("path must be a file"))
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
