// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func MoveFile(c *gin.Context) {
	sourcePath := c.Query("source")
	destPath := c.Query("destination")

	if sourcePath == "" || destPath == "" {
		c.AbortWithError(400, errors.New("source and destination paths are required"))
		return
	}

	// Get absolute paths
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid source path"))
		return
	}

	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid destination path"))
		return
	}

	// Check if source exists
	sourceInfo, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(404, errors.New("source file not found"))
			return
		}
		c.AbortWithError(400, errors.New("unable to access source file"))
		return
	}

	// Check if destination parent directory exists
	destDir := filepath.Dir(absDestPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		c.AbortWithError(404, errors.New("destination directory does not exist"))
		return
	}

	// Check if destination already exists
	if _, err := os.Stat(absDestPath); err == nil {
		c.AbortWithError(404, errors.New("destination already exists"))
		return
	}

	// Perform the move operation
	err = os.Rename(absSourcePath, absDestPath)
	if err != nil {
		// If rename fails (e.g., across different devices), try copy and delete
		if err := copyFile(absSourcePath, absDestPath, sourceInfo); err != nil {
			c.AbortWithError(400, fmt.Errorf("failed to move file: %w", err))
			return
		}

		// If copy successful, delete the source
		if err := os.RemoveAll(absSourcePath); err != nil {
			// If delete fails, inform that the file was copied but not deleted
			c.JSON(http.StatusOK, gin.H{
				"message": "file copied successfully but source could not be deleted",
				"error":   fmt.Sprintf("failed to delete source: %v", err),
			})
			return
		}
	}

	c.Status(200)
}

func copyFile(src, dst string, srcInfo os.FileInfo) error {
	if srcInfo.IsDir() {
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Create relative path
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(dst, relPath)

			if info.IsDir() {
				return os.MkdirAll(targetPath, info.Mode())
			}

			// Copy the file
			return copyFileContents(path, targetPath, info.Mode())
		})
	}
	return copyFileContents(src, dst, srcInfo.Mode())
}

func copyFileContents(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = bufio.NewReader(in).WriteTo(out)
	return err
}
