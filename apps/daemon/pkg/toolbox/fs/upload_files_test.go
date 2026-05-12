// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Drives /files/bulk-upload end-to-end through the gin router and verifies the
// load-bearing behaviours: bytes land at the destination, the parent directory is
// created on demand, and (negatively) no leftover temp files are left behind.
func TestUploadFilesStreamsToDisk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()

	type want struct {
		path     string
		body     string
		mkParent bool
	}
	wants := []want{
		{path: filepath.Join(tempDir, "small.txt"), body: "hello, world"},
		{path: filepath.Join(tempDir, "empty.bin"), body: ""},
		// Ensure the parent dir is created on the fly.
		{path: filepath.Join(tempDir, "nested", "deep", "file.bin"), body: strings.Repeat("X", 64*1024), mkParent: true},
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for i, w := range wants {
		if err := mw.WriteField(formField(i, "path"), w.path); err != nil {
			t.Fatalf("write path field: %v", err)
		}
		part, err := mw.CreateFormFile(formField(i, "file"), filepath.Base(w.path))
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte(w.body)); err != nil {
			t.Fatalf("write file body: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/bulk-upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(rec)
	engine.POST("/files/bulk-upload", UploadFiles)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	for _, w := range wants {
		got, err := os.ReadFile(w.path)
		if err != nil {
			t.Fatalf("read %s: %v", w.path, err)
		}
		if string(got) != w.body {
			t.Fatalf("%s: expected %q, got %q", w.path, w.body, string(got))
		}
		// Atomic write should have left no temp file behind.
		entries, err := os.ReadDir(filepath.Dir(w.path))
		if err != nil {
			t.Fatalf("readdir %s: %v", filepath.Dir(w.path), err)
		}
		for _, e := range entries {
			if strings.Contains(e.Name(), ".daytona-upload-") {
				t.Fatalf("leftover temp file: %s", filepath.Join(filepath.Dir(w.path), e.Name()))
			}
		}
	}
}

// Asserts that errors during the file-body copy do not leave a partial file on disk
// at the destination — atomic-rename means we either commit a complete file or none.
func TestUploadFilesAtomicCleanupOnTruncatedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "partial.bin")

	// Hand-craft a multipart body that opens a file part but truncates mid-stream.
	// The reader will see an unexpected EOF inside io.Copy, surfacing as an error
	// for that part. The destination must NOT exist after the request.
	boundary := "DaytonaTestBoundary"
	body := &bytes.Buffer{}
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"files[0].path\"\r\n\r\n")
	body.WriteString(dest)
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"files[0].file\"; filename=\"partial.bin\"\r\n")
	body.WriteString("Content-Type: application/octet-stream\r\n\r\n")
	body.WriteString("partial-data-no-trailing-boundary")
	// Note: intentionally missing the closing "\r\n--BOUNDARY--\r\n".

	req := httptest.NewRequest(http.MethodPost, "/files/bulk-upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rec := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(rec)
	engine.POST("/files/bulk-upload", UploadFiles)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected destination not to exist after truncated upload, got err=%v", err)
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		t.Fatalf("expected empty directory, found leftover: %s", e.Name())
	}
}

// Empty path values are rejected with a per-index error; later parts are still
// processed, mirroring the existing bulk-upload contract of best-effort batching.
func TestUploadFilesRejectsEmptyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	if err := mw.WriteField("files[0].path", "  "); err != nil {
		t.Fatalf("write empty path: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/bulk-upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(rec)
	engine.POST("/files/bulk-upload", UploadFiles)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "empty path") {
		t.Fatalf("expected empty-path error, got %s", rec.Body.String())
	}
	if _, err := os.ReadDir(tempDir); err != nil {
		t.Fatalf("tempdir gone: %v", err)
	}
}

// Reading the request body via Request.Body proves the daemon doesn't depend on Gin
// having buffered the form into MaxMultipartMemory — we read the multipart envelope
// part-by-part. If anything ever switches it back to FormFile, this test will fail
// because the buffered parser would have consumed the body before our reader runs.
func TestUploadFilesUsesStreamingMultipartReader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "stream-only.bin")

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	if err := mw.WriteField("files[0].path", dest); err != nil {
		t.Fatalf("write field: %v", err)
	}
	part, err := mw.CreateFormFile("files[0].file", "stream-only.bin")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	payload := strings.Repeat("y", 1024*1024) // 1 MiB — well under MaxMultipartMemory
	if _, err := io.Copy(part, strings.NewReader(payload)); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/bulk-upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(rec)
	engine.POST("/files/bulk-upload", UploadFiles)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("payload mismatch: got %d bytes, want %d", len(got), len(payload))
	}
}

func formField(idx int, suffix string) string {
	switch idx {
	case 0:
		return "files[0]." + suffix
	case 1:
		return "files[1]." + suffix
	case 2:
		return "files[2]." + suffix
	}
	return ""
}
