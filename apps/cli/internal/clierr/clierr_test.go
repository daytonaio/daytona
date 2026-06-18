// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package clierr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil error", err: nil, want: 0},
		{name: "non-clierr error", err: errors.New("boom"), want: 1},
		{name: "usage category", err: clierr.New(clierr.CategoryUsage, "bad flag"), want: 2},
		{name: "timeout category", err: clierr.New(clierr.CategoryTimeout, "timed out"), want: 124},
		{name: "auth category", err: clierr.New(clierr.CategoryAuth, "unauthorized"), want: 1},
		{name: "not_found category", err: clierr.New(clierr.CategoryNotFound, "missing"), want: 1},
		{name: "conflict category", err: clierr.New(clierr.CategoryConflict, "exists"), want: 1},
		{name: "rate_limit category", err: clierr.New(clierr.CategoryRateLimit, "slow down"), want: 1},
		{name: "server category", err: clierr.New(clierr.CategoryServer, "oops"), want: 1},
		{name: "network category", err: clierr.New(clierr.CategoryNetwork, "refused"), want: 1},
		{name: "explicit code overrides category", err: clierr.New(clierr.CategoryUsage, "exec failed").WithCode(255), want: 255},
		{name: "explicit code on server category", err: clierr.New(clierr.CategoryServer, "exec failed").WithCode(7), want: 7},
		{name: "wrapped clierr is unwrapped", err: fmt.Errorf("context: %w", clierr.New(clierr.CategoryTimeout, "timed out")), want: 124},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clierr.ExitCode(tt.err); got != tt.want {
				t.Errorf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		wantCategory clierr.Category
		wantHint     string
	}{
		{name: "400 maps to usage", status: 400, wantCategory: clierr.CategoryUsage},
		{name: "401 maps to auth with login hint", status: 401, wantCategory: clierr.CategoryAuth, wantHint: "run 'daytona login' to reauthenticate"},
		{name: "403 maps to auth with permissions hint", status: 403, wantCategory: clierr.CategoryAuth, wantHint: "check that your API key has sufficient permissions for this action"},
		{name: "404 maps to not_found", status: 404, wantCategory: clierr.CategoryNotFound},
		{name: "409 maps to conflict", status: 409, wantCategory: clierr.CategoryConflict},
		{name: "429 maps to rate_limit", status: 429, wantCategory: clierr.CategoryRateLimit},
		{name: "500 maps to server", status: 500, wantCategory: clierr.CategoryServer},
		{name: "503 maps to server", status: 503, wantCategory: clierr.CategoryServer},
		{name: "unlisted 4xx defaults to server", status: 418, wantCategory: clierr.CategoryServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clierr.FromHTTPStatus(tt.status, "request failed")
			if err.Category != tt.wantCategory {
				t.Errorf("FromHTTPStatus(%d) category = %q, want %q", tt.status, err.Category, tt.wantCategory)
			}
			if err.Hint != tt.wantHint {
				t.Errorf("FromHTTPStatus(%d) hint = %q, want %q", tt.status, err.Hint, tt.wantHint)
			}
			if err.Message != "request failed" {
				t.Errorf("FromHTTPStatus(%d) message = %q, want %q", tt.status, err.Message, "request failed")
			}
		})
	}
}

func TestErrorHintComposition(t *testing.T) {
	tests := []struct {
		name string
		err  *clierr.Error
		want string
	}{
		{name: "no hint returns message only", err: clierr.New(clierr.CategoryServer, "request failed"), want: "request failed"},
		{name: "hint is appended with separator", err: clierr.New(clierr.CategoryAuth, "Unauthorized").WithHint("run 'daytona login' to reauthenticate"), want: "Unauthorized - run 'daytona login' to reauthenticate"},
		{name: "newf formats message", err: clierr.Newf(clierr.CategoryUsage, "invalid value %q", "x"), want: `invalid value "x"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithCodeOverride(t *testing.T) {
	err := clierr.New(clierr.CategoryTimeout, "timed out")
	if got := clierr.ExitCode(err); got != 124 {
		t.Fatalf("ExitCode before WithCode = %d, want 124", got)
	}

	if returned := err.WithCode(255); returned != err {
		t.Error("WithCode should return the receiver for chaining")
	}
	if got := clierr.ExitCode(err); got != 255 {
		t.Errorf("ExitCode after WithCode(255) = %d, want 255", got)
	}
}

func TestHasCategory(t *testing.T) {
	tests := []struct {
		name string
		err  error
		cat  clierr.Category
		want bool
	}{
		{name: "direct match", err: clierr.New(clierr.CategoryNotFound, "gone"), cat: clierr.CategoryNotFound, want: true},
		{name: "wrapped match", err: fmt.Errorf("context: %w", clierr.New(clierr.CategoryNotFound, "gone")), cat: clierr.CategoryNotFound, want: true},
		{name: "nil error", err: nil, cat: clierr.CategoryNotFound, want: false},
		{name: "wrong category", err: clierr.New(clierr.CategoryServer, "boom"), cat: clierr.CategoryNotFound, want: false},
		{name: "plain error", err: errors.New("boom"), cat: clierr.CategoryNotFound, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clierr.HasCategory(tt.err, tt.cat); got != tt.want {
				t.Errorf("HasCategory(%v, %q) = %v, want %v", tt.err, tt.cat, got, tt.want)
			}
		})
	}
}
