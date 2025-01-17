// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
)

func GetFileInfo(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	info, err := getFileInfo(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(404, errors.New("file not found"))
			return
		}
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, info)
}
