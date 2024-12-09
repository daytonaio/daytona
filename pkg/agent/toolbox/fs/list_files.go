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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		info, err := GetFileInfo(filepath.Join(path, file.Name()))
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	c.JSON(200, fileInfos)
}
