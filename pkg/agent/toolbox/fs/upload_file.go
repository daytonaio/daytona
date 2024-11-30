// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.Status(200)
}
