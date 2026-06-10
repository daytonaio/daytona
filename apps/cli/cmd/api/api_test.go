// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

func assertUsageError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *clierr.Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *clierr.Error, got %T: %v", err, err)
	}
	if cliErr.Category != clierr.CategoryUsage {
		t.Errorf("expected category %q, got %q", clierr.CategoryUsage, cliErr.Category)
	}
}

func TestJoinURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		path string
		want string
	}{
		{name: "no slashes", base: "https://api.daytona.io", path: "sandbox", want: "https://api.daytona.io/sandbox"},
		{name: "trailing slash on base", base: "https://api.daytona.io/", path: "sandbox", want: "https://api.daytona.io/sandbox"},
		{name: "leading slash on path", base: "https://api.daytona.io", path: "/sandbox", want: "https://api.daytona.io/sandbox"},
		{name: "both slashes", base: "https://api.daytona.io/", path: "/sandbox", want: "https://api.daytona.io/sandbox"},
		{name: "base with path segment", base: "https://app.daytona.io/api", path: "/sandbox/abc", want: "https://app.daytona.io/api/sandbox/abc"},
		{name: "path with query string", base: "https://api.daytona.io", path: "/sandbox?verbose=true", want: "https://api.daytona.io/sandbox?verbose=true"},
		{name: "empty path", base: "https://api.daytona.io", path: "", want: "https://api.daytona.io/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinURL(tt.base, tt.path); got != tt.want {
				t.Errorf("joinURL(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}

func TestNormalizeMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		want    string
		wantErr bool
	}{
		{name: "uppercase passthrough", method: "GET", want: "GET"},
		{name: "lowercase normalized", method: "post", want: "POST"},
		{name: "mixed case normalized", method: "DeLeTe", want: "DELETE"},
		{name: "surrounding whitespace trimmed", method: " put ", want: "PUT"},
		{name: "patch allowed", method: "patch", want: "PATCH"},
		{name: "head allowed", method: "HEAD", want: "HEAD"},
		{name: "unsupported method rejected", method: "OPTIONS", wantErr: true},
		{name: "garbage rejected", method: "FOO", wantErr: true},
		{name: "empty rejected", method: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeMethod(tt.method)
			if tt.wantErr {
				assertUsageError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("normalizeMethod(%q) unexpected error: %v", tt.method, err)
			}
			if got != tt.want {
				t.Errorf("normalizeMethod(%q) = %q, want %q", tt.method, got, tt.want)
			}
		})
	}
}

func TestResolveBody(t *testing.T) {
	bodyFile := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(bodyFile, []byte(`{"name":"my-sandbox"}`), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		method   string
		input    string
		stdin    string
		want     string
		wantBody bool
		wantErr  bool
	}{
		{name: "no input means no body", method: http.MethodGet, input: ""},
		{name: "no input with POST means no body", method: http.MethodPost, input: ""},
		{name: "stdin body for POST", method: http.MethodPost, input: "-", stdin: `{"a":1}`, want: `{"a":1}`, wantBody: true},
		{name: "file body for PUT", method: http.MethodPut, input: bodyFile, want: `{"name":"my-sandbox"}`, wantBody: true},
		{name: "file body for PATCH", method: http.MethodPatch, input: bodyFile, want: `{"name":"my-sandbox"}`, wantBody: true},
		{name: "empty stdin body still counts as body", method: http.MethodPost, input: "-", stdin: "", want: "", wantBody: true},
		{name: "input with GET rejected", method: http.MethodGet, input: bodyFile, wantErr: true},
		{name: "stdin input with GET rejected", method: http.MethodGet, input: "-", wantErr: true},
		{name: "input with HEAD rejected", method: http.MethodHead, input: bodyFile, wantErr: true},
		{name: "input with DELETE rejected", method: http.MethodDelete, input: "-", wantErr: true},
		{name: "missing file errors", method: http.MethodPost, input: filepath.Join(t.TempDir(), "missing.json"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, hasBody, err := resolveBody(tt.method, tt.input, strings.NewReader(tt.stdin))
			if tt.wantErr {
				assertUsageError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("resolveBody(%q, %q) unexpected error: %v", tt.method, tt.input, err)
			}
			if hasBody != tt.wantBody {
				t.Errorf("resolveBody(%q, %q) hasBody = %v, want %v", tt.method, tt.input, hasBody, tt.wantBody)
			}
			if string(body) != tt.want {
				t.Errorf("resolveBody(%q, %q) body = %q, want %q", tt.method, tt.input, body, tt.want)
			}
		})
	}
}

func TestTrailingNewlineWriter(t *testing.T) {
	tests := []struct {
		name   string
		writes []string
		want   string
	}{
		{name: "no output stays empty", writes: nil, want: ""},
		{name: "missing newline appended", writes: []string{`{"ok":true}`}, want: "{\"ok\":true}\n"},
		{name: "existing newline kept", writes: []string{"line\n"}, want: "line\n"},
		{name: "tracks last chunk", writes: []string{"a\n", "b"}, want: "a\nb\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := &trailingNewlineWriter{w: &buf}
			for _, chunk := range tt.writes {
				if _, err := w.Write([]byte(chunk)); err != nil {
					t.Fatal(err)
				}
			}
			if err := w.finish(); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
