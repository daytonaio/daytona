// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newComputerUseJSONContext(t *testing.T, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal test request: %v", err)
	}

	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")

	return ctx, recorder
}

func TestWrapClickHandlerReturnsBadRequestForValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := WrapClickHandler(func(req *MouseClickRequest) (*MouseClickResponse, error) {
		if req.Button == "wheel" {
			return nil, errors.New("unsupported mouse button")
		}

		return &MouseClickResponse{Position: Position{X: req.X, Y: req.Y}}, nil
	})

	ctx, recorder := newComputerUseJSONContext(t, "/computeruse/mouse/click", map[string]any{
		"x":      100,
		"y":      200,
		"button": "wheel",
		"double": false,
	})

	handler(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "unsupported mouse button") {
		t.Fatalf("expected validation error in response body, got %q", recorder.Body.String())
	}
}

func TestWrapScrollHandlerReturnsBadRequestForValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := WrapScrollHandler(func(req *MouseScrollRequest) (*ScrollResponse, error) {
		if req.Direction == "left" {
			return nil, errors.New("unsupported scroll direction")
		}

		return &ScrollResponse{Success: true}, nil
	})

	ctx, recorder := newComputerUseJSONContext(t, "/computeruse/mouse/scroll", map[string]any{
		"x":         10,
		"y":         20,
		"direction": "left",
		"amount":    1,
	})

	handler(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "unsupported scroll direction") {
		t.Fatalf("expected validation error in response body, got %q", recorder.Body.String())
	}
}
