// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type fileResult struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

// UploadFiles godoc
//
//	@Summary		Upload multiple files
//	@Description	Upload multiple files with their destination paths
//	@Tags			file-system
//	@Accept			multipart/form-data
//	@Success		200
//	@Router			/files/bulk-upload [post]
//
//	@id				UploadFiles
func UploadFiles(c *gin.Context) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid multipart form"}})
		return
	}

	dests := make(map[string]string)
	var errs []string
	var files []fileResult

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs = append(errs, fmt.Sprintf("reading part: %v", err))
			continue
		}

		name := part.FormName()

		if strings.HasSuffix(name, ".path") {
			data, err := io.ReadAll(part)
			if err != nil {
				idx := extractIndex(name)
				errs = append(errs, fmt.Sprintf("path[%s]: %v", idx, err))
				continue
			}
			idx := extractIndex(name)
			dest := strings.TrimSpace(string(data))
			if dest == "" {
				errs = append(errs, fmt.Sprintf("path[%s]: empty path", idx))
				continue
			}
			dests[idx] = dest
			continue
		}

		if strings.HasSuffix(name, ".file") {
			idx := extractIndex(name)
			dest, ok := dests[idx]
			if !ok {
				errs = append(errs, fmt.Sprintf("file[%s]: missing .path metadata", idx))
				continue
			}

			if d := filepath.Dir(dest); d != "" {
				if err := os.MkdirAll(d, 0o755); err != nil {
					errs = append(errs, fmt.Sprintf("%s: mkdir %s: %v", dest, d, err))
					continue
				}
			}

			f, err := os.Create(dest)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: create: %v", dest, err))
				continue
			}

			n, err := io.Copy(f, part)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: write: %v", dest, err))
			}
			f.Close()
			files = append(files, fileResult{Path: dest, Bytes: n})
			continue
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs, "files": files})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func extractIndex(fieldName string) string {
	s := strings.TrimPrefix(fieldName, "files[")
	return strings.TrimSuffix(strings.TrimSuffix(s, "].path"), "].file")
}
