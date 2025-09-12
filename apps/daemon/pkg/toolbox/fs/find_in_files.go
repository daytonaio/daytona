// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bufio"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// FindInFiles godoc
//
//	@Summary		Find text in files
//	@Description	Search for text pattern within files in a directory
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query	string	true	"Directory path to search in"
//	@Param			pattern	query	string	true	"Text pattern to search for"
//	@Success		200		{array}	Match
//	@Router			/files/find [get]
//
//	@id				FindInFiles
func FindInFiles(c *gin.Context) {
	path := c.Query("path")
	pattern := c.Query("pattern")
	if path == "" || pattern == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("path and pattern are required"))
		return
	}

	var matches []Match = make([]Match, 0)
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil {
			return nil
		}

		for i := 0; i < n; i++ {
			// skip binary files
			if buf[i] == 0 {
				return nil
			}
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return nil
		}

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), pattern) {
				matches = append(matches, Match{
					File:    filePath,
					Line:    lineNum,
					Content: scanner.Text(),
				})
			}
			lineNum++
		}
		return nil
	})

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, matches)
}
