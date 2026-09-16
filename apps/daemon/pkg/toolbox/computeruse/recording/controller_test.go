// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	recordingservice "github.com/daytonaio/daemon/pkg/recording"
)

func setupTestController(t *testing.T) (*RecordingController, *recordingservice.RecordingService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := recordingservice.NewRecordingService(logger, tempDir)
	controller := NewRecordingController(svc)
	return controller, svc
}

func TestListRecordings(t *testing.T) {
	controller, _ := setupTestController(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/computeruse/recordings", nil)

	controller.ListRecordings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("ListRecordings() status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp ListRecordingsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Recordings) != 0 {
		t.Errorf("len(resp.Recordings) = %d, want 0", len(resp.Recordings))
	}
}

func TestGetRecording_NotFound(t *testing.T) {
	controller, _ := setupTestController(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = []gin.Param{{Key: "id", Value: "nonexistent-id"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/computeruse/recordings/nonexistent-id", nil)

	controller.GetRecording(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GetRecording() status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestStopRecording_NotFound(t *testing.T) {
	controller, _ := setupTestController(t)

	body := bytes.NewBufferString(`{"id":"nonexistent-id"}`)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/computeruse/recordings/stop", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.StopRecording(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("StopRecording() status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestStartRecording_InvalidLabel(t *testing.T) {
	controller, _ := setupTestController(t)

	// Test with label
	body := bytes.NewBufferString(`{"label":"../bad-path"}`)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/computeruse/recordings/start", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.StartRecording(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("StartRecording() status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] != "invalid_label" {
		t.Errorf("resp[\"error\"] = %q, want \"invalid_label\"", resp["error"])
	}

	// Test with name (documentation mismatch support)
	bodyName := bytes.NewBufferString(`{"name":"../bad-path"}`)
	recorderName := httptest.NewRecorder()
	ctxName, _ := gin.CreateTestContext(recorderName)
	ctxName.Request = httptest.NewRequest(http.MethodPost, "/computeruse/recordings/start", bodyName)
	ctxName.Request.Header.Set("Content-Type", "application/json")

	controller.StartRecording(ctxName)

	if recorderName.Code != http.StatusBadRequest {
		t.Fatalf("StartRecording() with name status = %d, want %d", recorderName.Code, http.StatusBadRequest)
	}
	var respName map[string]string
	if err := json.Unmarshal(recorderName.Body.Bytes(), &respName); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if respName["error"] != "invalid_label" {
		t.Errorf("respName[\"error\"] = %q, want \"invalid_label\"", respName["error"])
	}
}
