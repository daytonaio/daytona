// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateFolder(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	// Get the permission mode from query params, default to 0755
	mode := c.Query("mode")
	var perm os.FileMode = 0755
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			c.AbortWithError(400, errors.New("invalid mode format"))
			return
		}
		perm = os.FileMode(modeNum)
	}

	if err := os.MkdirAll(path, perm); err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(201)
}
