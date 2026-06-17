// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/daytonaio/daemon/pkg/toolbox/fs"
)

func TestFilesRoutesAcceptSlashAndNoSlash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	fsController := r.Group("/files")
	{
		fsController.GET("/", fs.ListFiles)
		fsController.GET("", fs.ListFiles)
	}

	tempDir := t.TempDir()

	check := func(path string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	statusNoSlash := check("/files?path=" + tempDir)
	statusSlash := check("/files/?path=" + tempDir)

	if statusNoSlash != http.StatusOK {
		t.Fatalf("GET /files returned status %d, want %d", statusNoSlash, http.StatusOK)
	}

	if statusSlash != http.StatusOK {
		t.Fatalf("GET /files/ returned status %d, want %d", statusSlash, http.StatusOK)
	}
}

func TestDeleteFilesRoutesAcceptSlashAndNoSlash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	fsController := r.Group("/files")
	{
		fsController.DELETE("/", fs.DeleteFile)
		fsController.DELETE("", fs.DeleteFile)
	}

	tempDir := t.TempDir()

	writeTempFile := func(name string) string {
		p := filepath.Join(tempDir, name)
		if err := os.WriteFile(p, []byte("content"), 0o644); err != nil {
			t.Fatalf("failed to create temp file %q: %v", p, err)
		}
		return p
	}

	check := func(path string) int {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	fileNoSlash := writeTempFile("no-slash.txt")
	fileSlash := writeTempFile("slash.txt")

	statusNoSlash := check("/files?path=" + url.QueryEscape(fileNoSlash))
	statusSlash := check("/files/?path=" + url.QueryEscape(fileSlash))

	// Guard against the 307 trailing-slash redirect regression (#5029): the
	// route must be served directly, never via a 3xx redirect. Check this first
	// so a reintroduced redirect produces a clear failure message.
	if statusNoSlash >= 300 && statusNoSlash < 400 {
		t.Fatalf("DELETE /files was redirected with status %d, want a direct response", statusNoSlash)
	}

	if statusSlash >= 300 && statusSlash < 400 {
		t.Fatalf("DELETE /files/ was redirected with status %d, want a direct response", statusSlash)
	}

	if statusNoSlash != http.StatusNoContent {
		t.Fatalf("DELETE /files returned status %d, want %d", statusNoSlash, http.StatusNoContent)
	}

	if statusSlash != http.StatusNoContent {
		t.Fatalf("DELETE /files/ returned status %d, want %d", statusSlash, http.StatusNoContent)
	}
}

func TestNewServerConstructs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s := NewServer(ServerConfig{
		Logger:    logger,
		WorkDir:   t.TempDir(),
		SandboxId: "test-sandbox",
		ConfigDir: t.TempDir(),
	})

	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}
