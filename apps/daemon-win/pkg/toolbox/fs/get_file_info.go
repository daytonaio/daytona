// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetFileInfo godoc
//
//	@Summary		Get file information
//	@Description	Get detailed information about a file or directory
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query		string	true	"File or directory path"
//	@Success		200		{object}	FileInfo
//	@Router			/files/info [get]
//
//	@id				GetFileInfo
func GetFileInfo(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path is required"))
		return
	}

	info, err := getFileInfo(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if os.IsPermission(err) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, info)
}

func getFileInfo(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	// On Windows, we don't have Unix-style UID/GID
	// We use "SYSTEM" or "BUILTIN" as placeholder values
	owner := "SYSTEM"
	group := "BUILTIN\\Users"

	// Try to get owner information from Windows security descriptor
	// For simplicity, we use static values as getting actual Windows
	// file ownership requires Windows-specific APIs
	return FileInfo{
		Name:        info.Name(),
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime().String(),
		IsDir:       info.IsDir(),
		Owner:       owner,
		Group:       group,
		Permissions: fmt.Sprintf("%04o", info.Mode().Perm()),
	}, nil
}
