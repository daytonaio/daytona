// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

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

// Streams requested files (and per-file errors) as multipart/form-data.
// Honors client cancellations via Context.
func DownloadFiles(c *gin.Context) {
	var req struct {
		Paths []string `json:"paths"`
	}
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

	for _, rawPath := range req.Paths {
		absPath, err := filepath.Abs(rawPath)
		if err != nil || !fileExists(absPath) {
			writeErrorPart(c, mw, absPath, fmt.Sprintf("file not found or invalid: %s", absPath))
			continue
		}

		f, err := os.Open(absPath)
		if err != nil {
			writeErrorPart(c, mw, absPath, fmt.Sprintf("error opening file: %v", err))
			continue
		}

		if err := writeFilePart(c.Request.Context(), mw, absPath, f); err != nil {
			writeErrorPart(c, mw, absPath,
				fmt.Sprintf("error streaming file: %v", err))
		}
		f.Close()
	}
}

// Streams a file part using io.Copy and respects context cancellation.
func writeFilePart(ctx context.Context, mw *multipart.Writer, absPath string, r io.Reader) error {
	ctype := mime.TypeByExtension(filepath.Ext(absPath))
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", ctype)
	hdr.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", absPath),
	)

	part, err := mw.CreatePart(hdr)
	if err != nil {
		return err
	}

	// Wrap part with context-aware writer
	cw := &ctxWriter{ctx: ctx, w: part}
	_, err = io.Copy(cw, r)
	return err
}

// Writes an error message as a multipart part.
func writeErrorPart(ctx *gin.Context, mw *multipart.Writer, absPath, text string) {
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "text/plain; charset=utf-8")
	hdr.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "error", absPath),
	)
	if part, err := mw.CreatePart(hdr); err == nil {
		_, err := io.WriteString(part, text)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
