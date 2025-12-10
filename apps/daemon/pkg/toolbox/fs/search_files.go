// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gin-gonic/gin"
)

// SearchFiles godoc
//
//	@Summary		Search files by pattern
//	@Description	Search for files matching a specific pattern in a directory. Supports standard glob patterns including * (match within directory), ** (recursive directory traversal), ? (single character wildcard), and {a,b} (group patterns).
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query		string	true	"Directory path to search in"
//	@Param			pattern	query		string	true	"Glob pattern to match (e.g., *.txt, **/*.go, src/**/*.ts)"
//	@Success		200		{object}	SearchFilesResponse
//	@Router			/files/search [get]
//
//	@id				SearchFiles
func SearchFiles(c *gin.Context) {
	basePath := c.Query("path")
	pattern := c.Query("pattern")
	if basePath == "" || pattern == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path and pattern are required"))
		return
	}

	// Initialize matches as an empty slice (not nil) to ensure JSON returns [] instead of null
	matches := []string{}

	// Walk through the directory tree and match files using doublestar glob patterns
	err := filepath.Walk(basePath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		// Skip directories from results (only return files)
		if info.IsDir() {
			return nil
		}

		// Get the relative path from the base path for pattern matching
		relativePath, relErr := filepath.Rel(basePath, currentPath)
		if relErr != nil {
			return nil
		}

		// Use forward slashes for consistent glob pattern matching across platforms
		relativePathForMatch := filepath.ToSlash(relativePath)

		// Match the pattern against the relative path using doublestar
		// This supports **, *, ?, and {a,b} glob patterns
		matched, matchErr := doublestar.Match(pattern, relativePathForMatch)
		if matchErr != nil {
			// If pattern is invalid, try matching against just the filename for backward compatibility
			matched, _ = doublestar.Match(pattern, info.Name())
		}

		if matched {
			matches = append(matches, currentPath)
		}
		return nil
	})

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, SearchFilesResponse{
		Files: matches,
	})
}
