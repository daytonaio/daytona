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
	enableFullDuplex(c)

	reader, err := c.Request.MultipartReader()
	if err != nil {
		drainBody(c)
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
			break
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

			n, copyErr := io.Copy(f, part)
			if copyErr != nil {
				errs = append(errs, fmt.Sprintf("%s: write: %v", dest, copyErr))
			}
			// Inspect Close() — on FUSE-backed filesystems (e.g. mount-s3 for
			// volume mounts) the actual remote write/CompleteMultipartUpload
			// happens here, so a swallowed close error means silent data loss.
			if closeErr := f.Close(); closeErr != nil && copyErr == nil {
				errs = append(errs, fmt.Sprintf("%s: close: %v", dest, closeErr))
				continue
			}
			if copyErr != nil {
				continue
			}
			files = append(files, fileResult{Path: dest, Bytes: n})
			continue
		}
	}

	drainBody(c)

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs, "files": files})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// drainBody consumes remaining request body bytes before a response is sent.
// Go's net/http server only auto-drains up to 256KB; anything beyond that
// causes it to close the connection before the reverse proxy finishes reading
// the response, producing 'unexpected EOF' on the proxy side.
func drainBody(c *gin.Context) {
	_, _ = io.Copy(io.Discard, c.Request.Body)
}

// enableFullDuplex allows the handler to write a response while the client is
// still sending the request body. Without this, Go's HTTP/1.1 server may RST
// the connection before the reverse proxy finishes reading the response.
func enableFullDuplex(c *gin.Context) {
	rc := http.NewResponseController(c.Writer)
	_ = rc.EnableFullDuplex()
}

func extractIndex(fieldName string) string {
	s := strings.TrimPrefix(fieldName, "files[")
	return strings.TrimSuffix(strings.TrimSuffix(s, "].path"), "].file")
}
