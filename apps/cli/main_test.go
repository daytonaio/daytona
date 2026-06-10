// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func TestExecuteErrorPayload(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantError string
		wantCode  string
		wantHint  string
	}{
		{
			name:      "clierr maps message, category, and hint",
			err:       clierr.New(clierr.CategoryAuth, "unauthorized").WithHint("run 'daytona login' to reauthenticate"),
			wantError: "unauthorized",
			wantCode:  "auth",
			wantHint:  "run 'daytona login' to reauthenticate",
		},
		{
			name:      "clierr without hint",
			err:       clierr.New(clierr.CategoryNotFound, "sandbox not found"),
			wantError: "sandbox not found",
			wantCode:  "not_found",
		},
		{
			name:      "wrapped clierr is unwrapped",
			err:       fmt.Errorf("context: %w", clierr.New(clierr.CategoryUsage, "bad flag").WithHint("see --help")),
			wantError: "bad flag",
			wantCode:  "usage",
			wantHint:  "see --help",
		},
		{
			name:      "plain error maps to generic code",
			err:       errors.New("boom"),
			wantError: "boom",
			wantCode:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := executeErrorPayload(tt.err)
			if payload.Error != tt.wantError {
				t.Errorf("Error = %q, want %q", payload.Error, tt.wantError)
			}
			if payload.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", payload.Code, tt.wantCode)
			}
			if payload.Hint != tt.wantHint {
				t.Errorf("Hint = %q, want %q", payload.Hint, tt.wantHint)
			}
		})
	}
}

func TestExecuteErrorPayloadJSONShape(t *testing.T) {
	t.Run("hint key omitted when empty", func(t *testing.T) {
		data, err := json.Marshal(executeErrorPayload(errors.New("boom")))
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		if got := string(data); got != `{"error":"boom","code":"error"}` {
			t.Errorf("payload JSON = %s, want {\"error\":\"boom\",\"code\":\"error\"}", got)
		}
	})

	t.Run("hint key present for clierr with hint", func(t *testing.T) {
		payload := executeErrorPayload(clierr.New(clierr.CategoryAuth, "unauthorized").WithHint("log in"))
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		if !strings.Contains(string(data), `"hint":"log in"`) {
			t.Errorf("payload JSON = %s, missing hint key", data)
		}
	})
}
