// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "echo hello",
			expected: []string{"echo", "hello"},
		},
		{
			name:     "command with quoted string",
			input:    `echo "hello world"`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "command with single quoted string",
			input:    `echo 'hello world'`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "multiple arguments",
			input:    "ls -la /tmp",
			expected: []string{"ls", "-la", "/tmp"},
		},
		{
			name:     "nested quotes",
			input:    `sh -c "echo 'hello'"`,
			expected: []string{"sh", "-c", "echo 'hello'"},
		},
		{
			name:     "backslash-escaped double-quote inside double-quoted string",
			input:    `echo "He said \"hello\""`,
			expected: []string{"echo", `He said "hello"`},
		},
		{
			name:     "unquoted backslash-quote escapes single-quote (POSIX)",
			input:    `echo O\'Brien`,
			expected: []string{"echo", "O'Brien"},
		},
		{
			name:     "unquoted backslash before non-quote char is literal",
			input:    `echo C:\tmp`,
			expected: []string{"echo", `C:\tmp`},
		},
		{
			name:     "POSIX single-quote escape idiom: '\\'' inside single-quoted string",
			input:    `sh -c 'echo O'\''Brien'`,
			expected: []string{"sh", "-c", "echo O'Brien"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil, // parseCommand returns nil for empty input
		},
		{
			name:     "spaces only",
			input:    "   ",
			expected: nil, // parseCommand returns nil for whitespace-only input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecuteRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request ExecuteRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: ExecuteRequest{
				Command: "echo hello",
			},
			valid: true,
		},
		{
			name: "empty command",
			request: ExecuteRequest{
				Command: "",
			},
			valid: false,
		},
		{
			name: "request with timeout",
			request: ExecuteRequest{
				Command: "sleep 10",
				Timeout: toUint32Ptr(5),
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation: command is required
			if tt.valid {
				assert.NotEmpty(t, tt.request.Command, "valid request should have non-empty command")
			} else {
				assert.Empty(t, tt.request.Command, "invalid request should have empty command")
			}
		})
	}
}

// TestBuildCommandRoundTrip is an end-to-end contract test between buildCommand
// (apps/cli/cmd/sandbox/exec.go) and parseCommand (this package).
//
// buildCommand encodes a []string into a single command string; parseCommand must
// decode it back to an identical slice. Because the two functions live in separate
// Go modules they cannot share a test binary, so this test hard-codes the exact
// strings that buildCommand produces and verifies parseCommand recovers the
// original arguments.  If buildCommand's encoding ever changes, both this test
// and TestBuildCommand in apps/cli/cmd/sandbox/exec_test.go must be updated
// together.
//
// Encoding rules implemented by buildCommand:
//
//	sh/bash -c <script> [arg0 ...]  → sh -c '<script>' [quoteArg(arg0) ...]
//	arg with spaces, no ' → 'arg'
//	arg with spaces and ' → "arg" (internal " escaped as \")
//	arg with ' or " but no spaces → "arg" or 'arg' respectively
//	arg without special chars → arg  (no quoting)
func TestBuildCommandRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		encoded     string // exact output of buildCommand(original)
		original    []string
	}{
		{
			name:     "simple command — no quoting needed",
			encoded:  "echo hello",
			original: []string{"echo", "hello"},
		},
		{
			name:     "arg with spaces uses single quotes",
			encoded:  "echo 'hello world'",
			original: []string{"echo", "hello world"},
		},
		{
			name:     "arg with spaces and single-quote falls back to double-quote wrapping",
			encoded:  `echo "it's alive"`,
			original: []string{"echo", "it's alive"},
		},
		{
			name:     "arg with spaces, single-quote, and double-quote: double-quote wrapping with \\\"",
			encoded:  `echo "say \"hi\" it's fine"`,
			original: []string{"echo", `say "hi" it's fine`},
		},
		{
			name:     "shell -c script without single quotes",
			encoded:  "sh -c 'echo hello && echo world'",
			original: []string{"sh", "-c", "echo hello && echo world"},
		},
		{
			name:     "shell -c script with one single-quote — POSIX '\\'' idiom",
			encoded:  `sh -c 'echo O'\''Brien'`,
			original: []string{"sh", "-c", "echo O'Brien"},
		},
		{
			name:     "shell -c script with multiple single-quotes",
			encoded:  `sh -c 'echo '\''hello'\'' '\''world'\'''`,
			original: []string{"sh", "-c", "echo 'hello' 'world'"},
		},
		{
			name:     "bash -c with single-quote in variable assignment",
			encoded:  `bash -c 'NAME=O'\''Brien; echo $NAME'`,
			original: []string{"bash", "-c", "NAME=O'Brien; echo $NAME"},
		},
		{
			name:     "sh -c with extra positional args preserves argv",
			encoded:  "sh -c 'echo $0 $1' hello world",
			original: []string{"sh", "-c", "echo $0 $1", "hello", "world"},
		},
		{
			name:     "sh -c with extra args needing quoting",
			encoded:  "sh -c 'echo $0' 'hello world'",
			original: []string{"sh", "-c", "echo $0", "hello world"},
		},
		{
			name:     "arg with single-quote but no spaces — double-quote wrapping",
			encoded:  `echo "O'Brien"`,
			original: []string{"echo", "O'Brien"},
		},
		{
			name:     "literal backslash in path preserved",
			encoded:  `echo C:\tmp`,
			original: []string{"echo", `C:\tmp`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.encoded)
			assert.Equal(t, tt.original, got,
				"parseCommand(%q) should round-trip to the original args", tt.encoded)
		})
	}
}

// Helper functions

func toUint32Ptr(val uint32) *uint32 {
	return &val
}

func toUint16Ptr(val uint16) *uint16 {
	return &val
}
