// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"os"

	"github.com/gin-gonic/gin"
)

func DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(400, gin.H{"error": "path is required"})
		return
	}

	// Check if recursive deletion is requested
	recursive := c.Query("recursive") == "true"

	// Get file info to check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// If it's a directory and recursive flag is not set, return error
	if info.IsDir() && !recursive {
		c.JSON(400, gin.H{"error": "cannot delete directory without recursive flag"})
		return
	}

	var deleteErr error
	if recursive {
		deleteErr = os.RemoveAll(path)
	} else {
		deleteErr = os.Remove(path)
	}

	if deleteErr != nil {
		c.JSON(500, gin.H{"error": deleteErr.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "file deleted successfully"})
}
