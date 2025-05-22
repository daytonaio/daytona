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
