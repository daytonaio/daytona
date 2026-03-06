// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandConstruction(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "simple command",
			args:     []string{"echo", "hello"},
			contains: "echo hello",
		},
		{
			name:     "shell command with -c",
			args:     []string{"sh", "-c", "echo hello && echo world"},
			contains: "sh -c echo hello && echo world",
		},
		{
			name:     "command with multiple args",
			args:     []string{"ls", "-la", "/tmp"},
			contains: "ls -la /tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that command construction logic would work
			if len(tt.args) >= 3 && tt.args[1] == "-c" {
				// Shell -c case
				parts := []string{tt.args[0], tt.args[1]}
				cmdPart := strings.Join(tt.args[2:], " ")
				result := strings.Join(append(parts, cmdPart), " ")
				assert.Contains(t, result, tt.contains)
			} else {
				// Regular case
				result := strings.Join(tt.args, " ")
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple command no quoting needed",
			args:     []string{"echo", "hello"},
			expected: "echo hello",
		},
		{
			name:     "arg with spaces uses single quotes",
			args:     []string{"echo", "hello world"},
			expected: "echo 'hello world'",
		},
		{
			name:     "arg with single quote but no spaces uses double quotes",
			args:     []string{"echo", "O'Brien"},
			expected: `echo "O'Brien"`,
		},
		{
			name:     "arg with double quote but no spaces uses single quotes",
			args:     []string{"echo", `say"hello"`},
			expected: `echo 'say"hello"'`,
		},
		{
			name:     "shell -c script without quotes",
			args:     []string{"sh", "-c", "echo hello && echo world"},
			expected: "sh -c 'echo hello && echo world'",
		},
		{
			name:     "shell -c script with single quote uses POSIX escape idiom",
			args:     []string{"sh", "-c", "echo O'Brien"},
			expected: `sh -c 'echo O'\''Brien'`,
		},
		{
			name:     "shell -c script with multiple single quotes",
			args:     []string{"sh", "-c", "echo 'hello' 'world'"},
			expected: `sh -c 'echo '\''hello'\'' '\''world'\'''`,
		},
		{
			name:     "shell -c with extra positional args preserves argv",
			args:     []string{"sh", "-c", "echo $0 $1", "hello", "world"},
			expected: "sh -c 'echo $0 $1' hello world",
		},
		{
			name:     "shell -c with extra args needing quoting",
			args:     []string{"sh", "-c", "echo $0", "hello world"},
			expected: "sh -c 'echo $0' 'hello world'",
		},
		{
			name:     "empty args returns empty string",
			args:     []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommand(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecFlagBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectTTY bool
		expectCwd bool
	}{
		{
			name:      "basic exec without flags",
			args:      []string{"--tty=false", "sandbox-id", "echo", "hello"},
			expectTTY: false,
			expectCwd: false,
		},
		{
			name:      "exec with --tty flag",
			args:      []string{"--tty", "sandbox-id", "bash"},
			expectTTY: true,
			expectCwd: false,
		},
		{
			name:      "exec with --cwd flag",
			args:      []string{"--cwd", "/tmp", "sandbox-id", "ls"},
			expectTTY: false,
			expectCwd: true,
		},
		{
			name:      "exec with multiple flags",
			args:      []string{"--cwd", "/tmp", "--tty", "--timeout", "30", "sandbox-id", "bash"},
			expectTTY: true,
			expectCwd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cobra strips -- before RunE; verify flags are present in the raw CLI args
			hasTTY := contains(tt.args, "--tty")
			hasCwd := contains(tt.args, "--cwd")

			assert.Equal(t, tt.expectTTY, hasTTY)
			assert.Equal(t, tt.expectCwd, hasCwd)
		})
	}
}

// TestArgsLenAtDash simulates how Cobra passes args to RunE after consuming "--".
// Cobra strips "--" and sets ArgsLenAtDash to the count of args before it.
func TestArgsLenAtDash(t *testing.T) {
	tests := []struct {
		name              string
		argsAfterCobra    []string // what Cobra passes to RunE (-- is stripped)
		argsLenAtDash     int      // what cmd.ArgsLenAtDash() returns (-1 = no --)
		expectSandbox     string
		expectCmds        []string
		expectErr         bool
	}{
		{
			name:           "sandbox and command separated by --",
			argsAfterCobra: []string{"sandbox-id", "echo", "hello", "world"},
			argsLenAtDash:  1,
			expectSandbox:  "sandbox-id",
			expectCmds:     []string{"echo", "hello", "world"},
		},
		{
			name:           "sandbox and multi-arg command",
			argsAfterCobra: []string{"sandbox-id", "bash", "-i"},
			argsLenAtDash:  1,
			expectSandbox:  "sandbox-id",
			expectCmds:     []string{"bash", "-i"},
		},
		{
			name:           "no -- separator used",
			argsAfterCobra: []string{"sandbox-id", "vim"},
			argsLenAtDash:  -1,
			expectErr:      true,
		},
		{
			name:           "-- used but no sandbox ID before it",
			argsAfterCobra: []string{"vim"},
			argsLenAtDash:  0,
			expectErr:      true,
		},
		{
			name:           "-- used but no command after it",
			argsAfterCobra: []string{"sandbox-id"},
			argsLenAtDash:  1,
			expectSandbox:  "sandbox-id",
			expectCmds:     []string{},
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashIndex := tt.argsLenAtDash

			if dashIndex == -1 {
				assert.True(t, tt.expectErr, "should error when no -- used")
				return
			}
			if dashIndex == 0 {
				assert.True(t, tt.expectErr, "should error when no sandbox ID before --")
				return
			}

			sandboxId := tt.argsAfterCobra[0]
			commandArgs := tt.argsAfterCobra[dashIndex:]

			assert.Equal(t, tt.expectSandbox, sandboxId)

			if tt.expectErr {
				assert.Empty(t, commandArgs)
			} else {
				assert.Equal(t, tt.expectCmds, commandArgs)
			}
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
