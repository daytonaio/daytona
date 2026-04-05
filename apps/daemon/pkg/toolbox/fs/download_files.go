// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

// Wraps an io.Writer and aborts writes if the context is canceled.
type ctxWriter struct {
	ctx context.Context
	w   io.Writer
}

func (cw *ctxWriter) Write(p []byte) (int, error) {
	select {
	case <-cw.ctx.Done():
		return 0, cw.ctx.Err()
	default:
	}
	return cw.w.Write(p)
}

// DownloadFiles godoc
//
//	@Summary		Download multiple files
//	@Description	Download multiple files by providing their paths. Successful files are returned as multipart parts named `file`. Per-file failures are returned as multipart parts named `error` with JSON payloads shaped like ErrorResponse.
//	@Tags			file-system
//	@Accept			json
//	@Produce		multipart/form-data
//	@Param			downloadFiles	body		FilesDownloadRequest	true	"Paths of files to download"
//	@Success		200				{object}	gin.H					"Multipart response with file parts and JSON error parts"
//	@Router			/files/bulk-download [post]
//
//	@id				DownloadFiles
func DownloadFiles(c *gin.Context) {
	var req FilesDownloadRequest
	if err := c.BindJSON(&req); err != nil || len(req.Paths) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "request body must be {\"paths\": [ ... ]} and non-empty",
		})
		return
	}

	const boundary = "DAYTONA-FILE-BOUNDARY"
	c.Status(http.StatusOK)
	c.Header("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	mw := multipart.NewWriter(c.Writer)
	if err := mw.SetBoundary(boundary); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to set multipart boundary",
		})
		return
	}
	defer mw.Close() // ensure final boundary is written

	for _, path := range req.Paths {
		downloadErr := classifyDownloadPathError(c, path)
		if downloadErr != nil {
			writeErrorPart(c, mw, path, *downloadErr)
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			writeErrorPart(c, mw, path, classifyOpenFileError(c, path, err))
			continue
		}

		if err := writeFilePart(c.Request.Context(), mw, path, f); err != nil {
			f.Close()

			// If streaming fails after the multipart file part has started, emitting a
			// second error part for the same path breaks the response contract.
			c.Error(err)
			return
		}
		f.Close()
	}
}

// Streams a file part using io.Copy and respects context cancellation.
func writeFilePart(ctx context.Context, mw *multipart.Writer, path string, r io.Reader) error {
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", ctype)
	hdr.Set("Content-Disposition", multipartContentDisposition("file", path))

	part, err := mw.CreatePart(hdr)
	if err != nil {
		return err
	}

	// Wrap part with context-aware writer
	cw := &ctxWriter{ctx: ctx, w: part}
	_, err = io.Copy(cw, r)
	return err
}

// Writes a structured error response as a multipart part.
// Pre-marshals to a buffer so a failed encode cannot corrupt a half-written part.
func writeErrorPart(ctx *gin.Context, mw *multipart.Writer, path string, errorResponse common.ErrorResponse) {
	payload, err := json.Marshal(errorResponse)
	if err != nil {
		ctx.Error(err)
		return
	}

	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "application/json; charset=utf-8")
	hdr.Set("Content-Disposition", multipartContentDisposition("error", path))

	part, err := mw.CreatePart(hdr)
	if err != nil {
		ctx.Error(err)
		return
	}
	if _, err := part.Write(payload); err != nil {
		ctx.Error(err)
	}
}

func classifyDownloadPathError(ctx *gin.Context, path string) *common.ErrorResponse {
	// Validate the input path before touching the filesystem.
	if path == "" {
		errorResponse := newFileDownloadErrorResponse(
			ctx,
			path,
			http.StatusBadRequest,
			"INVALID_FILE_PATH",
			"invalid file path: path is empty",
		)
		return &errorResponse
	}

	info, err := os.Stat(path)
	if err != nil {
		statusCode, errorCode, message := classifyPathStatError(path, err)
		errorResponse := newFileDownloadErrorResponse(ctx, path, statusCode, errorCode, message)
		return &errorResponse
	}

	if info.IsDir() {
		errorResponse := newFileDownloadErrorResponse(
			ctx,
			path,
			http.StatusBadRequest,
			"INVALID_FILE_PATH",
			fmt.Sprintf("invalid file path: path points to a directory: %s", path),
		)
		return &errorResponse
	}

	return nil
}

func multipartContentDisposition(formName string, path string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\r", "", "\n", "").Replace(path)
	return fmt.Sprintf(`form-data; name="%s"; filename="%s"`, formName, escaped)
}

func classifyPathStatError(path string, err error) (int, string, string) {
	// Preserve a specific not-found classification for missing files.
	if errors.Is(err, os.ErrNotExist) {
		return http.StatusNotFound, "FILE_NOT_FOUND", fmt.Sprintf("file not found: %s", path)
	}

	if errors.Is(err, os.ErrPermission) {
		return http.StatusForbidden, "FILE_ACCESS_DENIED", fmt.Sprintf("permission denied: %s", path)
	}

	if errors.Is(err, os.ErrInvalid) {
		return http.StatusBadRequest, "INVALID_FILE_PATH", fmt.Sprintf("invalid file path: %s", path)
	}

	return http.StatusInternalServerError, "FILE_READ_FAILED", fmt.Sprintf("failed to access file: %v", err)
}

func classifyOpenFileError(ctx *gin.Context, path string, err error) common.ErrorResponse {
	statusCode, errorCode, message := classifyPathStatError(path, err)
	// Use a more specific fallback message for open errors.
	if errorCode == "FILE_READ_FAILED" {
		message = fmt.Sprintf("failed to open file: %v", err)
	}
	return newFileDownloadErrorResponse(ctx, path, statusCode, errorCode, message)
}

func newFileDownloadErrorResponse(
	ctx *gin.Context,
	path string,
	statusCode int,
	code string,
	message string,
) common.ErrorResponse {
	return common.ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Timestamp:  time.Now().UTC(),
		Path:       path,
		Method:     ctx.Request.Method,
	}
}
