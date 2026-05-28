// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func newRouter(source string, defaultHandler func(*gin.Context, error) ErrorResponse) *gin.Engine {
	r := gin.New()
	r.Use(NewErrorMiddleware(source, defaultHandler))
	return r
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) (int, ErrorResponse) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var body ErrorResponse
	if rr.Body.Len() > 0 {
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response body: %v\nbody: %s", err, rr.Body.String())
		}
	}
	return rr.Code, body
}

func TestMiddleware_HTTPError_EmitsFullEnvelope(t *testing.T) {
	r := newRouter("DAYTONA_DAEMON", nil)
	r.GET("/x", func(c *gin.Context) {
		c.Error(NewNotFoundError(errors.New("widget")))
	})

	status, body := doRequest(t, r, http.MethodGet, "/x")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
	if body.StatusCode != http.StatusNotFound {
		t.Errorf("body.StatusCode = %d, want %d", body.StatusCode, http.StatusNotFound)
	}
	if body.Source != "DAYTONA_DAEMON" {
		t.Errorf("body.Source = %q, want DAYTONA_DAEMON", body.Source)
	}
	if body.Code != "NOT_FOUND" {
		t.Errorf("body.Code = %q, want NOT_FOUND", body.Code)
	}
	if body.Path != "/x" {
		t.Errorf("body.Path = %q, want /x", body.Path)
	}
	if body.Method != http.MethodGet {
		t.Errorf("body.Method = %q, want GET", body.Method)
	}
	if body.Timestamp.IsZero() {
		t.Error("body.Timestamp must be set")
	}
}

func TestMiddleware_DefaultHandler_FallbackPath(t *testing.T) {
	r := newRouter("DAYTONA_RUNNER", nil)
	r.GET("/y", func(c *gin.Context) {
		c.Error(errors.New("uncategorised boom"))
	})

	status, body := doRequest(t, r, http.MethodGet, "/y")

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", status, http.StatusInternalServerError)
	}
	if body.Source != "DAYTONA_RUNNER" {
		t.Errorf("body.Source = %q, want DAYTONA_RUNNER", body.Source)
	}
	if body.Code != "" {
		t.Errorf("body.Code = %q, want empty", body.Code)
	}
	if body.Message != "uncategorised boom" {
		t.Errorf("body.Message = %q, want %q", body.Message, "uncategorised boom")
	}
}

func TestMiddleware_CustomHandler_IsUsedForNonHTTPErrors(t *testing.T) {
	calls := 0
	custom := func(ctx *gin.Context, err error) ErrorResponse {
		calls++
		return NewErrorResponseForCtx(ctx, http.StatusTeapot, "DAYTONA_API", "custom: "+err.Error())
	}
	r := newRouter("DAYTONA_API", custom)
	r.GET("/z", func(c *gin.Context) {
		c.Error(errors.New("oops"))
	})

	status, body := doRequest(t, r, http.MethodGet, "/z")

	if calls != 1 {
		t.Errorf("custom handler calls = %d, want 1", calls)
	}
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}
	if body.Source != "DAYTONA_API" {
		t.Errorf("body.Source = %q, want DAYTONA_API", body.Source)
	}
	if body.Message != "custom: oops" {
		t.Errorf("body.Message = %q, want %q", body.Message, "custom: oops")
	}
}

func TestMiddleware_CustomHandler_NotCalledForHTTPErrors(t *testing.T) {
	calls := 0
	custom := func(ctx *gin.Context, err error) ErrorResponse {
		calls++
		return NewErrorResponseForCtx(ctx, http.StatusInternalServerError, "DAYTONA_API", err.Error())
	}
	r := newRouter("DAYTONA_API", custom)
	r.GET("/q", func(c *gin.Context) {
		c.Error(NewConflictError(errors.New("dup")))
	})

	status, body := doRequest(t, r, http.MethodGet, "/q")

	if calls != 0 {
		t.Errorf("custom handler calls = %d, want 0", calls)
	}
	if status != http.StatusConflict {
		t.Errorf("status = %d, want %d", status, http.StatusConflict)
	}
	if body.Code != "CONFLICT" {
		t.Errorf("body.Code = %q, want CONFLICT", body.Code)
	}
}

func TestMiddleware_NoErrors_DoesNothing(t *testing.T) {
	r := newRouter("DAYTONA_DAEMON", nil)
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got, want := rr.Body.String(), `{"hello":"world"}`; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestMiddleware_AlreadyWritten_DoesNotOverride(t *testing.T) {
	r := newRouter("DAYTONA_DAEMON", nil)
	r.GET("/w", func(c *gin.Context) {
		c.String(http.StatusOK, "raw-body")
		c.Error(NewBadRequestError(errors.New("late")))
	})

	req := httptest.NewRequest(http.MethodGet, "/w", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got, want := rr.Body.String(), "raw-body"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestNewErrorResponseFromHTTPError_BadRequestExtraction(t *testing.T) {
	r := newRouter("DAYTONA_RUNNER", nil)
	r.GET("/u", func(c *gin.Context) {
		c.Error(NewBadRequestError(errors.New("docker: unable to find user nobody: no entry found")))
	})

	status, body := doRequest(t, r, http.MethodGet, "/u")

	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if body.Code != "BAD_REQUEST" {
		t.Errorf("body.Code = %q, want BAD_REQUEST", body.Code)
	}
	if body.Message != "bad request: unable to find user nobody" {
		t.Errorf("body.Message = %q, want %q", body.Message, "bad request: unable to find user nobody")
	}
}
