// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"

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
