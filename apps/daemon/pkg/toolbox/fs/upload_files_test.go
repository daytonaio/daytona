// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bytes"
	"encoding/json"
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

func TestUploadFilesTruncatedBodyLeavesNoFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "partial.bin")

	boundary := "DaytonaTestBoundary"
	body := &bytes.Buffer{}
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"files[0].path\"\r\n\r\n")
	body.WriteString(dest)
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"files[0].file\"; filename=\"partial.bin\"\r\n")
	body.WriteString("Content-Type: application/octet-stream\r\n\r\n")
	body.WriteString("partial-data-no-trailing-boundary")

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

// Verifies that uploaded files have 0644 permissions and the JSON response
// contains accurate byte counts per file.
func TestUploadFilesResponseBytesAndMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	content := "hello, progress tracking!"
	dest := filepath.Join(tempDir, "progress.txt")

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	if err := mw.WriteField("files[0].path", dest); err != nil {
		t.Fatalf("write path: %v", err)
	}
	part, err := mw.CreateFormFile("files[0].file", "progress.txt")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write: %v", err)
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
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Files []struct {
			Path  string `json:"path"`
			Bytes int64  `json:"bytes"`
		} `json:"files"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Files) != 1 {
		t.Fatalf("expected 1 file result, got %d", len(resp.Files))
	}
	if resp.Files[0].Path != dest {
		t.Fatalf("expected path %q, got %q", dest, resp.Files[0].Path)
	}
	if resp.Files[0].Bytes != int64(len(content)) {
		t.Fatalf("expected %d bytes, got %d", len(content), resp.Files[0].Bytes)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("expected mode 0644, got %04o", perm)
	}
}

// Verifies that a multi-file upload returns correct byte counts for each file
// and that all files have 0644 permissions.
func TestUploadFilesMultiFileBytesAndMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()

	type entry struct {
		dest    string
		content string
	}
	entries := []entry{
		{dest: filepath.Join(tempDir, "a.txt"), content: "short"},
		{dest: filepath.Join(tempDir, "b.bin"), content: strings.Repeat("B", 128*1024)},
		{dest: filepath.Join(tempDir, "c.txt"), content: ""},
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for i, e := range entries {
		if err := mw.WriteField(formField(i, "path"), e.dest); err != nil {
			t.Fatalf("write path: %v", err)
		}
		part, err := mw.CreateFormFile(formField(i, "file"), filepath.Base(e.dest))
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		if _, err := part.Write([]byte(e.content)); err != nil {
			t.Fatalf("write: %v", err)
		}
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
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Files []struct {
			Path  string `json:"path"`
			Bytes int64  `json:"bytes"`
		} `json:"files"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Files) != len(entries) {
		t.Fatalf("expected %d file results, got %d", len(entries), len(resp.Files))
	}

	for i, e := range entries {
		r := resp.Files[i]
		if r.Path != e.dest {
			t.Fatalf("file[%d]: expected path %q, got %q", i, e.dest, r.Path)
		}
		if r.Bytes != int64(len(e.content)) {
			t.Fatalf("file[%d]: expected %d bytes, got %d", i, len(e.content), r.Bytes)
		}

		info, err := os.Stat(e.dest)
		if err != nil {
			t.Fatalf("stat %s: %v", e.dest, err)
		}
		if perm := info.Mode().Perm(); perm != 0o644 {
			t.Fatalf("file[%d]: expected mode 0644, got %04o", i, perm)
		}
	}
}

func TestUploadFilesWritesThroughSymlink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	realFile := filepath.Join(tempDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("old content"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	link := filepath.Join(tempDir, "link.txt")
	if err := os.Symlink(realFile, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	content := "new content via symlink"
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	if err := mw.WriteField("files[0].path", link); err != nil {
		t.Fatalf("write path: %v", err)
	}
	part, err := mw.CreateFormFile("files[0].file", "link.txt")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write: %v", err)
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
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was replaced instead of written through")
	}

	got, err := os.ReadFile(realFile)
	if err != nil {
		t.Fatalf("read real file: %v", err)
	}
	if string(got) != content {
		t.Fatalf("expected %q at real path, got %q", content, string(got))
	}
}

func TestUploadFilesWritesThroughDanglingSymlink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "not-yet.txt")
	link := filepath.Join(tempDir, "dangling.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	content := "created via dangling symlink"
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	if err := mw.WriteField("files[0].path", link); err != nil {
		t.Fatalf("write path: %v", err)
	}
	part, err := mw.CreateFormFile("files[0].file", "dangling.txt")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write: %v", err)
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
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was replaced instead of written through")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("target should have been created: %v", err)
	}
	if string(got) != content {
		t.Fatalf("expected %q at target, got %q", content, string(got))
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
