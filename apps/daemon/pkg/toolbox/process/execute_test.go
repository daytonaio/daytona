// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.Default()
	r.POST("/process/execute", ExecuteCommand(logger))
	return r
}

func executeRequest(t *testing.T, router *gin.Engine, req ExecuteRequest) ExecuteResponse {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/process/execute", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ExecuteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return resp
}

func TestExecuteCommand_NotFound(t *testing.T) {
	router := setupRouter()
	resp := executeRequest(t, router, ExecuteRequest{
		Command: "nonexistent_command_xyz_12345",
	})

	if resp.ExitCode != 127 {
		t.Errorf("expected exit code 127 for command not found, got %d", resp.ExitCode)
	}
}

func TestExecuteCommand_Success(t *testing.T) {
	router := setupRouter()
	resp := executeRequest(t, router, ExecuteRequest{
		Command: "echo hello",
	})

	if resp.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", resp.ExitCode)
	}
	if resp.Result != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", resp.Result)
	}
}

func TestExecuteCommand_ExitCode(t *testing.T) {
	router := setupRouter()
	resp := executeRequest(t, router, ExecuteRequest{
		Command: "exit 42",
	})

	if resp.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", resp.ExitCode)
	}
}

func TestExecuteCommand_Pipe(t *testing.T) {
	router := setupRouter()
	resp := executeRequest(t, router, ExecuteRequest{
		Command: "echo hello | grep hello",
	})

	if resp.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", resp.ExitCode)
	}
	if resp.Result != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", resp.Result)
	}
}
