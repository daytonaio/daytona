// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"path/filepath"
	"testing"
)

func writeShellsFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "shells")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// unsetShellEnv unsets $SHELL for the test, restoring the original value
// (or unset state) afterwards via t.Setenv's cleanup.
func unsetShellEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SHELL", "")
	if err := os.Unsetenv("SHELL"); err != nil {
		t.Fatal(err)
	}
}

func resetShellCache(t *testing.T) {
	t.Helper()
	reset := func() {
		shellCacheMu.Lock()
		defer shellCacheMu.Unlock()
		cachedShell = ""
		shellCached = false
	}
	reset()
	t.Cleanup(reset)
}

func swapShellsFilePath(t *testing.T, path string) {
	t.Helper()
	original := shellsFilePath
	shellsFilePath = path
	t.Cleanup(func() { shellsFilePath = original })
}

func TestResolveShellPrefersZshOverBash(t *testing.T) {
	path := writeShellsFile(t, "/bin/sh\n/bin/bash\n/usr/bin/bash\n/bin/zsh\n/usr/bin/zsh\n")

	shell, ok := resolveShell(path)
	if !ok {
		t.Fatal("expected successful resolution")
	}
	if shell != "/usr/bin/zsh" {
		t.Fatalf("expected /usr/bin/zsh, got %q", shell)
	}
}

func TestResolveShellBashWhenNoZsh(t *testing.T) {
	path := writeShellsFile(t, "/bin/sh\n/bin/bash\n")

	shell, ok := resolveShell(path)
	if !ok {
		t.Fatal("expected successful resolution")
	}
	if shell != "/bin/bash" {
		t.Fatalf("expected /bin/bash, got %q", shell)
	}
}

func TestResolveShellEnvFallbackWhenNoPreferredShell(t *testing.T) {
	t.Setenv("SHELL", "/opt/custom/fish")
	path := writeShellsFile(t, "/bin/sh\n/opt/custom/fish\n")

	shell, ok := resolveShell(path)
	if !ok {
		t.Fatal("expected successful resolution")
	}
	if shell != "/opt/custom/fish" {
		t.Fatalf("expected /opt/custom/fish, got %q", shell)
	}
}

func TestResolveShellShWhenFileMissingAndNoShellEnv(t *testing.T) {
	unsetShellEnv(t)
	path := filepath.Join(t.TempDir(), "does-not-exist")

	shell, ok := resolveShell(path)
	if ok {
		t.Fatal("expected failed resolution for missing file")
	}
	if shell != "sh" {
		t.Fatalf("expected sh, got %q", shell)
	}
}

func TestResolveShellIgnoresCommentLines(t *testing.T) {
	unsetShellEnv(t)
	path := writeShellsFile(t, "# /usr/bin/zsh\n#/bin/zsh\n/bin/bash\n")

	shell, ok := resolveShell(path)
	if !ok {
		t.Fatal("expected successful resolution")
	}
	if shell != "/bin/bash" {
		t.Fatalf("expected /bin/bash, got %q", shell)
	}
}

func TestGetShellCachesSuccessfulResolution(t *testing.T) {
	resetShellCache(t)
	path := writeShellsFile(t, "/usr/bin/zsh\n")
	swapShellsFilePath(t, path)

	if shell := GetShell(); shell != "/usr/bin/zsh" {
		t.Fatalf("expected /usr/bin/zsh, got %q", shell)
	}

	// Mutate the file; the cached answer must not change.
	if err := os.WriteFile(path, []byte("/bin/bash\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if shell := GetShell(); shell != "/usr/bin/zsh" {
		t.Fatalf("expected cached /usr/bin/zsh, got %q", shell)
	}
}

func TestGetShellDoesNotCacheFailedResolution(t *testing.T) {
	resetShellCache(t)
	unsetShellEnv(t)
	path := filepath.Join(t.TempDir(), "shells")
	swapShellsFilePath(t, path)

	if shell := GetShell(); shell != "sh" {
		t.Fatalf("expected sh while file is missing, got %q", shell)
	}

	// The failed read must not have been cached: once the file appears,
	// the next call picks it up.
	if err := os.WriteFile(path, []byte("/bin/bash\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if shell := GetShell(); shell != "/bin/bash" {
		t.Fatalf("expected /bin/bash after file created, got %q", shell)
	}
}
