// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SearchFiles(c *gin.Context) {
	path := c.Query("path")
	pattern := c.Query("pattern")
	if path == "" || pattern == "" {
		c.AbortWithError(400, errors.New("path and pattern are required"))
		return
	}

	var matches []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		c.JSON(400, err)
		return
	}

	c.JSON(200, SearchFilesResponse{
		Files: matches,
	})
}
