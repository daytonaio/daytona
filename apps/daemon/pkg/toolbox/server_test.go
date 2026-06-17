// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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
		fsController.DELETE("/", fs.DeleteFile)
		fsController.DELETE("", fs.DeleteFile)
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

	deleteFileNoSlash, err := os.CreateTemp(t.TempDir(), "delete-no-slash")
	if err != nil {
		t.Fatal(err)
	}
	if err := deleteFileNoSlash.Close(); err != nil {
		t.Fatal(err)
	}

	deleteFileSlash, err := os.CreateTemp(t.TempDir(), "delete-slash")
	if err != nil {
		t.Fatal(err)
	}
	if err := deleteFileSlash.Close(); err != nil {
		t.Fatal(err)
	}

	checkDelete := func(path string) int {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	deleteStatusNoSlash := checkDelete("/files?path=" + deleteFileNoSlash.Name())
	deleteStatusSlash := checkDelete("/files/?path=" + deleteFileSlash.Name())

	if deleteStatusNoSlash != http.StatusNoContent {
		t.Fatalf("DELETE /files returned status %d, want %d", deleteStatusNoSlash, http.StatusNoContent)
	}

	if deleteStatusSlash != http.StatusNoContent {
		t.Fatalf("DELETE /files/ returned status %d, want %d", deleteStatusSlash, http.StatusNoContent)
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
