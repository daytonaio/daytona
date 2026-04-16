// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func downloadFileContext(t *testing.T, filePath string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/files/download?path="+url.QueryEscape(filePath), nil)
	DownloadFile(ctx)
	return recorder
}

func TestDownloadFileContentDisposition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ascii filename sets both filename and filename*", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "hello.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		recorder := downloadFileContext(t, filePath)

		got := recorder.Header().Get("Content-Disposition")
		want := `attachment; filename="hello.txt"; filename*=utf-8''hello.txt`
		if got != want {
			t.Errorf("Content-Disposition = %q, want %q", got, want)
		}
	})

	t.Run("unicode filename uses latin1 fallback and RFC 5987 encoding", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "日本語.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		recorder := downloadFileContext(t, filePath)

		got := recorder.Header().Get("Content-Disposition")
		if !strings.Contains(got, `filename="___.txt"`) {
			t.Errorf("expected latin1 fallback filename, got %q", got)
		}
		if !strings.Contains(got, `filename*=utf-8''%E6%97%A5%E6%9C%AC%E8%AA%9E.txt`) {
			t.Errorf("expected RFC 5987 encoded filename*, got %q", got)
		}
	})

	t.Run("filename with special characters is properly encoded", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "file (1).txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		recorder := downloadFileContext(t, filePath)

		got := recorder.Header().Get("Content-Disposition")
		if !strings.Contains(got, `filename*=utf-8''file%20%281%29.txt`) {
			t.Errorf("expected percent-encoded filename*, got %q", got)
		}
	})

	t.Run("control characters in filename are replaced", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "hello\tworld.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		recorder := downloadFileContext(t, filePath)

		got := recorder.Header().Get("Content-Disposition")
		if !strings.Contains(got, `filename="hello_world.txt"`) {
			t.Errorf("expected tab replaced with underscore in filename, got %q", got)
		}
	})

	t.Run("missing path returns 400", func(t *testing.T) {
		recorder := downloadFileContext(t, "")

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
		}
	})

	t.Run("nonexistent file returns 404", func(t *testing.T) {
		recorder := downloadFileContext(t, "/tmp/nonexistent-file-xyz.txt")

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
		}
	})

	t.Run("directory path returns 400", func(t *testing.T) {
		tempDir := t.TempDir()

		recorder := downloadFileContext(t, tempDir)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
		}
	})
}
