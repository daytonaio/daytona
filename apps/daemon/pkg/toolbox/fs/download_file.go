// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DownloadFile godoc
//
//	@Summary		Download a file
//	@Description	Download a file by providing its path
//	@Tags			file-system
//	@Produce		octet-stream
//	@Param			path	query	string	true	"File path to download"
//	@Success		200		{file}	binary
//	@Router			/files/download [get]
//
//	@id				DownloadFile
func DownloadFile(c *gin.Context) {
	requestedPath := c.Query("path")
	if requestedPath == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	requestedPath = filepath.Clean(requestedPath)

	if !filepath.IsAbs(requestedPath) {
		workDir, err := os.Getwd()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get working directory: %w", err))
			return
		}
		requestedPath = filepath.Join(workDir, requestedPath)
	}

	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid path: %w", err))
		return
	}

	evalPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if os.IsPermission(err) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid path: %w", err))
		return
	}
	absPath = evalPath

	workDir, err := os.Getwd()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get working directory: %w", err))
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get home directory: %w", err))
		return
	}

	if !strings.HasPrefix(absPath, workDir+string(filepath.Separator)) && absPath != workDir &&
		!strings.HasPrefix(absPath, homeDir+string(filepath.Separator)) && absPath != homeDir {
		c.AbortWithError(http.StatusForbidden, errors.New("access denied: path is outside of the allowed directory"))
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
