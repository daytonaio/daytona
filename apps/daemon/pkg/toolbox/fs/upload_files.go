// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

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
//
// UploadFiles streams multipart parts directly to disk without buffering. Each file body
// is io.Copy'd into a sibling temp file in the destination's directory, then renamed
// atomically onto the destination on success — readers never see a partial file. Any
// error mid-upload (including client disconnect via context cancellation) removes the
// temp file before returning.
//
// Wire format: a sequence of multipart parts alternating between
//   - text part named "files[<idx>].path" carrying the destination path, and
//   - file part named "files[<idx>].file" carrying the file body.
//
// The .path part for a given index MUST arrive before the corresponding .file part —
// the server does not buffer file bodies waiting for late metadata.
func UploadFiles(c *gin.Context) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid multipart form"}})
		return
	}

	dests := make(map[string]string)
	var errs []string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			// A non-EOF error means the multipart stream is no longer parseable
			// (truncated body, malformed boundary, etc.) — keep iterating would
			// loop forever, so record and stop.
			errs = append(errs, fmt.Sprintf("reading part: %v", err))
			break
		}

		name := part.FormName()
		idx := extractIndex(name)

		switch {
		case strings.HasSuffix(name, ".path"):
			data, readErr := io.ReadAll(part)
			if readErr != nil {
				errs = append(errs, fmt.Sprintf("path[%s]: %v", idx, readErr))
				continue
			}
			dest := strings.TrimSpace(string(data))
			if dest == "" {
				errs = append(errs, fmt.Sprintf("path[%s]: empty path", idx))
				continue
			}
			dests[idx] = dest

		case strings.HasSuffix(name, ".file"):
			dest, ok := dests[idx]
			if !ok {
				errs = append(errs, fmt.Sprintf("file[%s]: missing .path metadata", idx))
				continue
			}
			if err := writeUploadedPart(c.Request.Context(), part, dest); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", dest, err))
			}
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
		return
	}

	c.Status(http.StatusOK)
}

// writeUploadedPart atomically writes a single multipart file body to dest. Bytes are
// streamed into a temp file in the destination directory (so os.Rename is atomic on
// the same filesystem), then renamed over dest on success. Any failure — including
// context cancellation surfaced by ctxWriter — removes the temp file.
func writeUploadedPart(ctx context.Context, part *multipart.Part, dest string) error {
	if dir := filepath.Dir(dest); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".daytona-upload-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	cw := &ctxWriter{ctx: ctx, w: tmp}
	if _, err := io.Copy(cw, part); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	committed = true
	return nil
}

func extractIndex(fieldName string) string {
	s := strings.TrimPrefix(fieldName, "files[")
	return strings.TrimSuffix(strings.TrimSuffix(s, "].path"), "].file")
}
