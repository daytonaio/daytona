// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
)

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

	stat := info.Sys().(*syscall.Stat_t)
	return FileInfo{
		Name:        info.Name(),
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime().String(),
		IsDir:       info.IsDir(),
		Owner:       strconv.FormatUint(uint64(stat.Uid), 10),
		Group:       strconv.FormatUint(uint64(stat.Gid), 10),
		Permissions: fmt.Sprintf("%04o", info.Mode().Perm()),
	}, nil
}
