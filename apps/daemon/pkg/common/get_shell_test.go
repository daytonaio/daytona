// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"testing"
)

func TestSessionShellStartsWithNoInitFlags(t *testing.T) {
	args := GetShellArgs()
	if len(args) == 0 {
		t.Fatal("GetShellArgs() returned empty slice")
	}

	shell := args[0]
	switch shell {
	case "/usr/bin/zsh", "/bin/zsh":
		if len(args) < 2 {
			t.Fatalf("expected no-init flag for zsh %q, got only the shell path", shell)
		}
		if args[1] != "-f" {
			t.Fatalf("expected -f for zsh %q, got %q", shell, args[1])
		}
	case "/usr/bin/bash", "/bin/bash":
		if len(args) < 3 {
			t.Fatalf("expected --norc --noprofile for bash %q, got %v", shell, args)
		}
		if args[1] != "--norc" || args[2] != "--noprofile" {
			t.Fatalf("expected --norc --noprofile for bash %q, got %v", shell, args[1:])
		}
	default:
		// Other shells get no extra flags — just the shell path is acceptable.
		if len(args) != 1 {
			t.Fatalf("expected only shell path for %q, got %v", shell, args)
		}
	}
}
