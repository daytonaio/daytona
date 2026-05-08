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
	"strconv"
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
		f, info, downloadErr := openDownloadFile(c, path)
		if downloadErr != nil {
			writeErrorPart(c, mw, path, *downloadErr)
			continue
		}

		if err := writeFilePart(c.Request.Context(), mw, path, f, info.Size()); err != nil {
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
func writeFilePart(ctx context.Context, mw *multipart.Writer, path string, r io.Reader, size int64) error {
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", ctype)
	hdr.Set("Content-Disposition", multipartContentDisposition("file", path))
	hdr.Set("Content-Length", strconv.FormatInt(size, 10))

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

// openDownloadFile opens path for reading and validates it is a downloadable
// regular file. On success, the caller owns closing the returned file.
//
// Stat is performed against the open file descriptor rather than the path, so
// the size used for Content-Length and the bytes streamed afterwards reference
// the same inode and cannot diverge under a concurrent rename or replace.
func openDownloadFile(ctx *gin.Context, path string) (*os.File, os.FileInfo, *common.ErrorResponse) {
	if path == "" {
		errorResponse := newFileDownloadErrorResponse(
			ctx,
			path,
			http.StatusBadRequest,
			"INVALID_FILE_PATH",
			"invalid file path: path is empty",
		)
		return nil, nil, &errorResponse
	}

	f, err := os.Open(path)
	if err != nil {
		errorResponse := classifyOpenFileError(ctx, path, err)
		return nil, nil, &errorResponse
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		statusCode, errorCode, message := classifyPathStatError(path, err)
		errorResponse := newFileDownloadErrorResponse(ctx, path, statusCode, errorCode, message)
		return nil, nil, &errorResponse
	}

	if info.IsDir() {
		f.Close()
		errorResponse := newFileDownloadErrorResponse(
			ctx,
			path,
			http.StatusBadRequest,
			"INVALID_FILE_PATH",
			fmt.Sprintf("invalid file path: path points to a directory: %s", path),
		)
		return nil, nil, &errorResponse
	}

	return f, info, nil
}

func multipartContentDisposition(formName string, path string) string {
	return fmt.Sprintf(`form-data; name="%s"; filename="%s"; filename*=utf-8''%s`,
		formName, toLatin1(path), encodeRFC5987(path))
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

// toLatin1 sanitizes s for use in a quoted Content-Disposition filename= parameter.
// It escapes backslashes and double quotes, strips CR/LF/NUL, replaces C0 control
// characters (0x01-0x1F), DEL (0x7F), and non-Latin1 (>0xFF) runes with '_'.
func toLatin1(s string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\r", "", "\n", "", "\x00", "").Replace(s)
	var buf []byte
	for _, r := range escaped {
		if r > 0xFF {
			buf = append(buf, '_')
		} else if r < 0x20 || r == 0x7F {
			buf = append(buf, '_')
		} else {
			buf = append(buf, byte(r))
		}
	}
	return string(buf)
}

// encodeRFC5987 percent-encodes per RFC 5987 attr-char rules.
// Safe: alphanumerics and !#$&+-.^_`|~ — everything else is encoded.
func encodeRFC5987(s string) string {
	var buf []byte
	for _, b := range []byte(s) {
		if isAttrChar(b) {
			buf = append(buf, b)
		} else {
			buf = append(buf, fmt.Sprintf("%%%02X", b)...)
		}
	}
	return string(buf)
}

// isAttrChar reports whether b is in the RFC 5987 attr-char safe set.
func isAttrChar(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	}
	switch b {
	case '!', '#', '$', '&', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}
