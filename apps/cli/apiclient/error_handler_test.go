// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/apiclient"
)

func response(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestHandleErrorResponse_403_PermissionsHint(t *testing.T) {
	t.Run("reproduction case: 403 appends permissions hint", func(t *testing.T) {
		res := response(http.StatusForbidden, `{"error":"Forbidden"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "check that your API key has sufficient permissions for this action") {
			t.Errorf("expected permissions hint in error, got: %q", err.Error())
		}
	})

	t.Run("403 with message field appends both message and hint", func(t *testing.T) {
		res := response(http.StatusForbidden, `{"error":"Forbidden","message":"snapshot push requires write access"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "snapshot push requires write access") {
			t.Errorf("expected server message in error, got: %q", msg)
		}
		if !strings.Contains(msg, "check that your API key has sufficient permissions for this action") {
			t.Errorf("expected permissions hint in error, got: %q", msg)
		}
	})

	t.Run("403 with array message field appends hint", func(t *testing.T) {
		res := response(http.StatusForbidden, `{"error":"Forbidden","message":["scope missing: snapshots"]}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "check that your API key has sufficient permissions for this action") {
			t.Errorf("expected permissions hint in error, got: %q", err.Error())
		}
	})

	t.Run("403 with empty error field falls back to raw body and appends hint", func(t *testing.T) {
		res := response(http.StatusForbidden, `{"error":"","message":"forbidden"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "check that your API key has sufficient permissions for this action") {
			t.Errorf("expected permissions hint in error, got: %q", err.Error())
		}
	})
}

func TestHandleErrorResponse_UnchangedBehavior(t *testing.T) {
	t.Run("401 still appends reauthentication hint", func(t *testing.T) {
		res := response(http.StatusUnauthorized, `{"error":"Unauthorized"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "run 'daytona login' to reauthenticate") {
			t.Errorf("expected reauth hint in 401 error, got: %q", err.Error())
		}
	})

	t.Run("401 does not include permissions hint", func(t *testing.T) {
		res := response(http.StatusUnauthorized, `{"error":"Unauthorized"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if strings.Contains(err.Error(), "check that your API key has sufficient permissions") {
			t.Errorf("401 error should not contain permissions hint, got: %q", err.Error())
		}
	})

	t.Run("403 does not include reauthentication hint", func(t *testing.T) {
		res := response(http.StatusForbidden, `{"error":"Forbidden"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if strings.Contains(err.Error(), "run 'daytona login' to reauthenticate") {
			t.Errorf("403 error should not contain reauth hint, got: %q", err.Error())
		}
	})

	t.Run("500 returns error without any hint", func(t *testing.T) {
		res := response(http.StatusInternalServerError, `{"error":"Internal Server Error"}`)
		err := apiclient.HandleErrorResponse(res, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		if strings.Contains(msg, "daytona login") || strings.Contains(msg, "check that your API key") {
			t.Errorf("500 error should contain no hint, got: %q", msg)
		}
	})

	t.Run("nil response returns the original request error", func(t *testing.T) {
		err := apiclient.HandleErrorResponse(nil, io.ErrUnexpectedEOF)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("expected original error, got: %v", err)
		}
	})

	t.Run("2xx with client error returns the original error unchanged", func(t *testing.T) {
		res := response(http.StatusOK, `{}`)
		err := apiclient.HandleErrorResponse(res, io.ErrUnexpectedEOF)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("expected original error for 2xx, got: %v", err)
		}
	})
}
