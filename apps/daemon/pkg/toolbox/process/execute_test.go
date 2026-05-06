// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"strings"
	"testing"
)

func TestShellSingleQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "''"},
		{"plain", "echo hi", "'echo hi'"},
		{"with single quote", "echo 'hi'", `'echo '\''hi'\'''`},
		{"with dollar", "echo $HOME", "'echo $HOME'"},
		{"with backtick", "echo `whoami`", "'echo `whoami`'"},
		{"with backslash", `echo "\n"`, `'echo "\n"'`},
		{"with newline", "line1\nline2", "'line1\nline2'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shellSingleQuote(tt.in); got != tt.want {
				t.Fatalf("shellSingleQuote(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDetachedWrapperCommand(t *testing.T) {
	got := detachedWrapperCommand("/bin/bash", "tmux new-session -d -s s 'bash --login'")

	// Must include the setsid availability probe so a missing util-linux
	// produces a clear error instead of a bare exit code 127.
	if !strings.Contains(got, "command -v setsid") {
		t.Errorf("wrapper should probe for setsid; got:\n%s", got)
	}

	// Must use exec + setsid -f + redirected stdio so the request returns
	// quickly and the detached process does not inherit our pipes.
	wantSubstrings := []string{
		"exec setsid -f",
		"</dev/null >/dev/null 2>&1",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("wrapper missing %q; got:\n%s", s, got)
		}
	}

	// User command (with embedded single quotes) must round-trip through
	// the quoting unchanged when bash parses it.
	wantInner := `'tmux new-session -d -s s '\''bash --login'\'''`
	if !strings.Contains(got, wantInner) {
		t.Errorf("inner command not quoted as expected; want substring %q; got:\n%s", wantInner, got)
	}
}
