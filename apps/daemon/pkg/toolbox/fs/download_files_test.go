// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext(method string) *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, "/files/bulk-download", nil)
	return ctx
}

func TestClassifyDownloadPathError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects empty paths as validation errors", func(t *testing.T) {
		ctx := newTestContext(http.MethodPost)

		errorResponse := classifyDownloadPathError(ctx, "")

		if errorResponse == nil {
			t.Fatal("expected an error response")
		}

		if errorResponse.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, errorResponse.StatusCode)
		}

		if errorResponse.Code != "INVALID_FILE_PATH" {
			t.Fatalf("expected INVALID_FILE_PATH, got %s", errorResponse.Code)
		}

		if errorResponse.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, errorResponse.Method)
		}
	})

	t.Run("rejects directories as invalid file paths", func(t *testing.T) {
		ctx := newTestContext(http.MethodPost)
		tempDir := t.TempDir()

		errorResponse := classifyDownloadPathError(ctx, tempDir)

		if errorResponse == nil {
			t.Fatal("expected an error response")
		}

		if errorResponse.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, errorResponse.StatusCode)
		}

		if errorResponse.Code != "INVALID_FILE_PATH" {
			t.Fatalf("expected INVALID_FILE_PATH, got %s", errorResponse.Code)
		}

		if !strings.Contains(errorResponse.Message, "directory") {
			t.Fatalf("expected directory message, got %q", errorResponse.Message)
		}

		if errorResponse.Path != tempDir {
			t.Fatalf("expected path %q, got %q", tempDir, errorResponse.Path)
		}
	})
}

func TestClassifyPathStatError(t *testing.T) {
	t.Run("maps missing files to not found", func(t *testing.T) {
		statusCode, errorCode, message := classifyPathStatError("/tmp/missing.txt", os.ErrNotExist)

		if statusCode != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, statusCode)
		}

		if errorCode != "FILE_NOT_FOUND" {
			t.Fatalf("expected FILE_NOT_FOUND, got %s", errorCode)
		}

		if !strings.Contains(message, "file not found") {
			t.Fatalf("expected not found message, got %q", message)
		}
	})

	t.Run("maps permission failures to access denied", func(t *testing.T) {
		statusCode, errorCode, message := classifyPathStatError("/tmp/locked.txt", os.ErrPermission)

		if statusCode != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, statusCode)
		}

		if errorCode != "FILE_ACCESS_DENIED" {
			t.Fatalf("expected FILE_ACCESS_DENIED, got %s", errorCode)
		}

		if !strings.Contains(message, "permission denied") {
			t.Fatalf("expected permission message, got %q", message)
		}
	})
}

func TestWriteErrorPart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newTestContext(http.MethodPost)
	buffer := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buffer)
	sourcePath := filepath.Join("/tmp", "missing.txt")

	writeErrorPart(
		ctx,
		writer,
		sourcePath,
		newFileDownloadErrorResponse(ctx, sourcePath, http.StatusNotFound, "FILE_NOT_FOUND", "file not found"),
	)

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(buffer.Bytes()), writer.Boundary())
	part, err := reader.NextPart()
	if err != nil {
		t.Fatalf("failed to read multipart part: %v", err)
	}

	if got := part.FormName(); got != "error" {
		t.Fatalf("expected form name error, got %s", got)
	}

	if got := part.Header.Get("Content-Disposition"); !strings.Contains(got, `filename="`+sourcePath+`"`) {
		t.Fatalf("expected raw filename %q in content disposition, got %q", sourcePath, got)
	}

	if got := part.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var errorResponse struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		Method     string `json:"method"`
		Path       string `json:"path"`
		StatusCode int    `json:"statusCode"`
	}

	if err := json.NewDecoder(part).Decode(&errorResponse); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errorResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, errorResponse.StatusCode)
	}

	if errorResponse.Code != "FILE_NOT_FOUND" {
		t.Fatalf("expected FILE_NOT_FOUND, got %s", errorResponse.Code)
	}

	if errorResponse.Method != http.MethodPost {
		t.Fatalf("expected method %s, got %s", http.MethodPost, errorResponse.Method)
	}

	if errorResponse.Path != sourcePath {
		t.Fatalf("expected path %q, got %q", sourcePath, errorResponse.Path)
	}
}
