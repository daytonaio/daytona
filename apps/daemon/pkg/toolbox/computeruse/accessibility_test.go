// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

func classifyA11yErrorResponse(t *testing.T, err error) (int, map[string]string) {
	t.Helper()
	rec := httptest.NewRecorder()

	router := gin.New()
	router.Use(common_errors.NewErrorMiddleware("DAYTONA_DAEMON", nil))
	router.GET("/test", func(c *gin.Context) {
		c.Error(classifyA11yError(err))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(rec, req)

	var raw map[string]interface{}
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &raw); decodeErr != nil {
		t.Fatalf("response body was not valid JSON: %v (raw=%q)", decodeErr, rec.Body.String())
	}

	body := make(map[string]string)
	for k, v := range raw {
		body[k] = fmt.Sprintf("%v", v)
	}
	return rec.Code, body
}

// TestClassifyA11yError pins the status + code mapping that SDKs branch on.
// The plugin wraps sentinels with fmt.Errorf("%w: ctx", sentinel, ...), so we
// feed in wrapped errors here to mirror the real call shape.
func TestClassifyA11yError(t *testing.T) {
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
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "no accessible root -> 404",
			err:        fmt.Errorf("%s", a11yMsgNoAccessibleRoot),
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "action not supported -> 400",
			err:        fmt.Errorf("%s: press", a11yMsgActionNotSupported),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid scope -> 400",
			err:        fmt.Errorf("%s: bogus", a11yMsgInvalidScope),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid request -> 400",
			err:        fmt.Errorf("%s: invalid node id", a11yMsgInvalidRequest),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid request containing unavailable text still -> 400",
			err:        fmt.Errorf(`%s: invalid node id %q`, a11yMsgInvalidRequest, a11yMsgUnavailable),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid request containing node-not-found text still -> 400",
			err:        fmt.Errorf(`%s: invalid node id %q`, a11yMsgInvalidRequest, a11yMsgNodeNotFound),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid request containing action-not-supported text still -> 400",
			err:        fmt.Errorf(`%s: invalid node id %q`, a11yMsgInvalidRequest, a11yMsgActionNotSupported),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid request containing invalid-scope text still -> 400",
			err:        fmt.Errorf(`%s: invalid node id %q`, a11yMsgInvalidRequest, a11yMsgInvalidScope),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid scope containing unavailable text still -> 400 invalid scope",
			err:        fmt.Errorf(`%s: got %q`, a11yMsgInvalidScope, a11yMsgUnavailable),
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "unknown error -> 500 with no code",
			err:        errors.New("something unmapped blew up deep in the plugin"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := classifyA11yErrorResponse(t, tc.err)
			if status != tc.wantStatus {
				t.Errorf("status = %d, want %d", status, tc.wantStatus)
			}
			if body["code"] != tc.wantCode {
				t.Errorf("code = %q, want %q", body["code"], tc.wantCode)
			}
			if body["message"] == "" {
				t.Error(`"message" field was empty; SDKs expect a human-readable message alongside the code`)
			}
		})
	}
}
