// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func ListFiles(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "."
	}

	files, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(404, err)
			return
		}
		c.AbortWithError(400, err)
		return
	}

	var fileInfos []FileInfo = make([]FileInfo, 0)
	for _, file := range files {
		info, err := getFileInfo(filepath.Join(path, file.Name()))
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, *info)
	}

	c.JSON(200, fileInfos)
}
