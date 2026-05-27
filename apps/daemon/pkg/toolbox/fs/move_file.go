// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

// MoveFile godoc
//
//	@Summary		Move or rename file/directory
//	@Description	Move or rename a file or directory from source to destination
//	@Tags			file-system
//	@Param			source		query	string	true	"Source file or directory path"
//	@Param			destination	query	string	true	"Destination file or directory path"
//	@Success		200
//	@Failure		400	{object}	common.ErrorResponse
//	@Failure		403	{object}	common.ErrorResponse
//	@Failure		404	{object}	common.ErrorResponse
//	@Failure		409	{object}	common.ErrorResponse
//	@Router			/files/move [post]
//
//	@id				MoveFile
func MoveFile(c *gin.Context) {
	sourcePath := c.Query("source")
	destPath := c.Query("destination")

	if sourcePath == "" || destPath == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("source and destination paths are required")))
		return
	}

	// Get absolute paths
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		c.Error(common_errors.NewBadRequestError(errors.New("invalid source path")))
		return
	}

	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		c.Error(common_errors.NewBadRequestError(errors.New("invalid destination path")))
		return
	}

	// Check if source exists
	sourceInfo, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(common.NewFileNotFoundError(err.Error()))
			return
		}
		if os.IsPermission(err) {
			c.Error(common.NewFileAccessDeniedError(err.Error()))
			return
		}
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	// Check if destination parent directory exists
	destDir := filepath.Dir(absDestPath)
	_, err = os.Stat(destDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(common.NewFileNotFoundError(err.Error()))
			return
		}
		if os.IsPermission(err) {
			c.Error(common.NewFileAccessDeniedError(err.Error()))
			return
		}
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	// Check if destination already exists
	if _, err := os.Stat(absDestPath); err == nil {
		c.Error(common_errors.NewConflictError(errors.New("destination already exists")))
		return
	}

	// Perform the move operation
	err = os.Rename(absSourcePath, absDestPath)
	if err != nil {
		// If rename fails (e.g., across different devices), try copy and delete
		if err := copyFile(absSourcePath, absDestPath, sourceInfo); err != nil {
			c.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to move file: %s", err.Error())))
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

	c.Status(http.StatusOK)
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
