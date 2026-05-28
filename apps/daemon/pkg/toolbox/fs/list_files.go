// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"net/http"
	"os"
	"path/filepath"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

// ListFiles godoc
//
//	@Summary		List files and directories
//	@Description	List files and directories in the specified path
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query		string	false	"Directory path to list (defaults to working directory)"
//	@Success		200		{array}		FileInfo
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		403		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/files [get]
//
//	@id				ListFiles
func ListFiles(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "."
	}

	files, err := os.ReadDir(path)
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

	var fileInfos = make([]FileInfo, 0)
	for _, file := range files {
		info, err := getFileInfo(filepath.Join(path, file.Name()))
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	c.JSON(http.StatusOK, fileInfos)
}
