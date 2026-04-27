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

	"github.com/daytonaio/daytona/apps/daemon/pkg/toolbox/fs"
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