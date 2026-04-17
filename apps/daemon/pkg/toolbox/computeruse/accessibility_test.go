// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// writeA11yErrorResponse runs writeA11yError against a test gin context and
// returns the recorded status + decoded JSON body so each case can assert
// against the machine-readable error shape SDKs will consume.
func writeA11yErrorResponse(t *testing.T, err error) (int, map[string]string) {
	t.Helper()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	writeA11yError(ctx, err)

	var body map[string]string
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &body); decodeErr != nil {
		t.Fatalf("response body was not a flat JSON object: %v (raw=%q)", decodeErr, rec.Body.String())
	}
	return rec.Code, body
}

// TestWriteA11yError pins the status + code mapping that SDKs branch on.
// The plugin wraps sentinels with fmt.Errorf("%w: ctx", sentinel, ...), so we
// feed in wrapped errors here to mirror the real call shape.
func TestWriteA11yError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "a11y bus unavailable -> 503",
			err:        fmt.Errorf("%s: session bus: connection refused", a11yMsgUnavailable),
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "A11Y_UNAVAILABLE",
		},
		{
			name:       "node not found -> 404",
			err:        fmt.Errorf("%s: :1.42:/org/a11y/atspi/accessible/99", a11yMsgNodeNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "A11Y_NODE_NOT_FOUND",
		},
		{
			name:       "no accessible root -> 404",
			err:        fmt.Errorf("%s", a11yMsgNoAccessibleRoot),
			wantStatus: http.StatusNotFound,
			wantCode:   "A11Y_NO_ACCESSIBLE_ROOT",
		},
		{
			name:       "action not supported -> 400",
			err:        fmt.Errorf("%s: press", a11yMsgActionNotSupported),
			wantStatus: http.StatusBadRequest,
			wantCode:   "A11Y_ACTION_NOT_SUPPORTED",
		},
		{
			name:       "invalid scope -> 400",
			err:        fmt.Errorf("%s: bogus", a11yMsgInvalidScope),
			wantStatus: http.StatusBadRequest,
			wantCode:   "A11Y_INVALID_SCOPE",
		},
		{
			name:       "unknown error -> 500 with A11Y_INTERNAL",
			err:        errors.New("something unmapped blew up deep in the plugin"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "A11Y_INTERNAL",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := writeA11yErrorResponse(t, tc.err)
			if status != tc.wantStatus {
				t.Errorf("status = %d, want %d", status, tc.wantStatus)
			}
			if body["code"] != tc.wantCode {
				t.Errorf("code = %q, want %q", body["code"], tc.wantCode)
			}
			if body["error"] == "" {
				t.Error(`"error" field was empty; SDKs still expect a human-readable message alongside the code`)
			}
		})
	}
}
