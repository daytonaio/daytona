// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// DeleteFile godoc
//
//	@Summary		Delete a file or directory
//	@Description	Delete a file or directory at the specified path
//	@Tags			file-system
//	@Param			path		query	string	true	"File or directory path to delete"
//	@Param			recursive	query	boolean	false	"Enable recursive deletion for directories"
//	@Success		204
//	@Router			/files [delete]
//
//	@id				DeleteFile
func DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	// Check if recursive deletion is requested
	recursive := c.Query("recursive") == "true"

	// Get file info to check if it's a directory
	info, err := os.Stat(path)
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

	// If it's a directory and recursive flag is not set, return error
	if info.IsDir() && !recursive {
		c.AbortWithError(http.StatusBadRequest, errors.New("cannot delete directory without recursive flag"))
		return
	}

	var deleteErr error
	if recursive {
		deleteErr = os.RemoveAll(path)
	} else {
		deleteErr = os.Remove(path)
	}

	if deleteErr != nil {
		c.AbortWithError(http.StatusBadRequest, deleteErr)
		return
	}

	c.Status(http.StatusNoContent)
}
