// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetFilePermissions godoc
//
//	@Summary		Set file permissions
//	@Description	Set file permissions for a file or directory. On Windows, only mode changes are supported (owner/group are ignored).
//	@Tags			file-system
//	@Param			path	query	string	true	"File or directory path"
//	@Param			owner	query	string	false	"Owner (not supported on Windows - ignored)"
//	@Param			group	query	string	false	"Group (not supported on Windows - ignored)"
//	@Param			mode	query	string	false	"File mode in octal format (e.g., 0755)"
//	@Success		200
//	@Router			/files/permissions [post]
//
//	@id				SetFilePermissions
func SetFilePermissions(c *gin.Context) {
	path := c.Query("path")
	mode := c.Query("mode")

	// Note: owner and group parameters are ignored on Windows
	// Windows uses ACLs (Access Control Lists) instead of Unix-style permissions
	// Implementing full ACL support would require Windows-specific APIs

	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	// convert to absolute path and check existence
	absPath, err := filepath.Abs(path)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid path"))
		return
	}

	_, err = os.Stat(absPath)
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

	// handle mode change
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid mode format"))
			return
		}

		// On Windows, os.Chmod only affects the read-only attribute
		// Mode 0444 or similar (no write bits) makes file read-only
		// Any mode with write bits removes read-only attribute
		if err := os.Chmod(absPath, os.FileMode(modeNum)); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to change mode: %w", err))
			return
		}
	}

	c.Status(http.StatusOK)
}
