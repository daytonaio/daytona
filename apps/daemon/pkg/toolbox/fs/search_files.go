// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

// SearchFiles godoc
//
//	@Summary		Search files by pattern
//	@Description	Search for files matching a specific pattern in a directory
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query		string	true	"Directory path to search in"
//	@Param			pattern	query		string	true	"File pattern to match (e.g., *.txt, *.go)"
//	@Success		200		{object}	SearchFilesResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/files/search [get]
//
//	@id				SearchFiles
func SearchFiles(c *gin.Context) {
	path := c.Query("path")
	pattern := c.Query("pattern")
	if path == "" || pattern == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("path and pattern are required")))
		return
	}

	matches := []string{}
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
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	c.JSON(http.StatusOK, SearchFilesResponse{
		Files: matches,
	})
}
