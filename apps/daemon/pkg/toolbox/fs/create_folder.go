// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateFolder godoc
//
//	@Summary		Create a folder
//	@Description	Create a folder with the specified path and optional permissions
//	@Tags			file-system
//	@Accept			json
//	@Param			path	query	string	true	"Folder path to create"
//	@Param			mode	query	string	true	"Octal permission mode (default: 0755)"
//	@Success		201
//	@Router			/files/folder [post]
//
//	@id				CreateFolder
func CreateFolder(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	// Get the permission mode from query params, default to 0755
	mode := c.Query("mode")
	var perm os.FileMode = 0755
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid mode format"))
			return
		}
		perm = os.FileMode(modeNum)
	}

	if err := os.MkdirAll(path, perm); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusCreated)
}
