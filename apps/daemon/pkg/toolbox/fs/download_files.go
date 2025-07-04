// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DownloadFiles streams requested files (and per-file errors) as multipart/form-data.
// Each file is sent in its own part, with Content-Disposition name="<path>" and filename="<basename>".
// At the end, if any files failed, a final summary part "download-summary.error.txt" lists all failures.
func DownloadFiles(c *gin.Context) {
	// 1. Parse body
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := c.BindJSON(&req); err != nil || len(req.Paths) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "request body must be {\"paths\": [ ... ]} and non-empty",
		})
		return
	}

	// 2. Prepare multipart writer
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

	// 3. Stream each file or error part, collect failures
	var failures []string

	for _, rawPath := range req.Paths {
		abs, err := filepath.Abs(rawPath)
		if err != nil || !fileExists(abs) {
			failures = append(failures, fmt.Sprintf("%s: not found or invalid", rawPath))
			writeTextPart(mw, rawPath, fmt.Sprintf("file not found or invalid: %s", rawPath))
			continue
		}

		f, err := os.Open(abs)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", rawPath, err))
			writeTextPart(mw, rawPath, fmt.Sprintf("error opening file: %v", err))
			continue
		}

		if err := writeFilePart(mw, rawPath, f); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", rawPath, err))
			writeTextPart(mw, rawPath, fmt.Sprintf("error streaming file: %v", err))
		}
		f.Close()
	}

	// 4. If any failures, append a summary part
	if len(failures) > 0 {
		summary := "download errors:\n- " + strings.Join(failures, "\n- ")
		writeTextPart(mw, "download-summary.error.txt", summary)
	}
}

// writeFilePart writes a file as one multipart part.
func writeFilePart(mw *multipart.Writer, rawPath string, r io.Reader) error {
	name := rawPath
	filename := filepath.Base(rawPath)

	ctype := mime.TypeByExtension(filepath.Ext(filename))
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", ctype)
	hdr.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename),
	)

	part, err := mw.CreatePart(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, r)
	return err
}

// writeTextPart writes a small text part assigned to the given name.
func writeTextPart(mw *multipart.Writer, name, text string) {
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "text/plain; charset=utf-8")
	hdr.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "error", filepath.Base(name)),
	)
	if part, err := mw.CreatePart(hdr); err == nil {
		io.WriteString(part, text)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
