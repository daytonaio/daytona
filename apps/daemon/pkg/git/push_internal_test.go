// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"
)

var pushTestCreds = &http.BasicAuth{
	Username: "push-test-user-xyz",
	Password: "push-test-token-abc123",
}

func TestBuildPushArgs(t *testing.T) {
	got := buildPushArgs("/work-dir", "refs/heads/main")
	require.Equal(t, []string{
		"-C", "/work-dir",
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"push",
		"origin",
		"refs/heads/main:refs/heads/main",
		"--progress",
	}, got)
}

func TestBuildPushArgs_NeverEmbedsCredsInArgs(t *testing.T) {
	args := buildPushArgs("/work-dir", "refs/heads/main")

	for _, arg := range args {
		require.NotContains(t, arg, pushTestCreds.Username,
			"username leaked into push args: %q", arg)
		require.NotContains(t, arg, pushTestCreds.Password,
			"password leaked into push args: %q", arg)
	}
}

func TestBuildPushArgs_DisablesHooks(t *testing.T) {
	args := buildPushArgs("/work-dir", "refs/heads/main")

	found := false
	for i, arg := range args {
		if arg == "core.hooksPath=/dev/null" && i > 0 && args[i-1] == "-c" {
			found = true
			break
		}
	}
	require.True(t, found, "push args must disable hooks via core.hooksPath=/dev/null")
}

func TestBuildPushArgs_VerifiesTLS(t *testing.T) {
	// Push must NOT skip TLS verification (parity with go-git PushOptions,
	// which does not set InsecureSkipTLS). Skipping verify would be a MITM
	// risk for the basic-auth token.
	args := buildPushArgs("/work-dir", "refs/heads/main")
	for _, arg := range args {
		require.NotEqual(t, "http.sslVerify=false", arg,
			"push args must NOT disable TLS verification")
	}
}

func TestPushCLI_EnvContainsAskpassAndCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", pushTestCreds)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
	require.Contains(t, env, "GIT_USERNAME="+pushTestCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+pushTestCreds.Password)
}

func TestPushCLI_EnvOmitsCredsWhenNil(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")

	for _, kv := range env {
		require.NotContains(t, kv, "GIT_USERNAME=",
			"GIT_USERNAME must not appear when auth is nil")
		require.NotContains(t, kv, "GIT_PASSWORD=",
			"GIT_PASSWORD must not appear when auth is nil")
	}
}

// initTestRepoOnBranch creates a fresh git repo on the given branch with one
// commit. Skips the test if `git` is unavailable.
func initTestRepoOnBranch(t *testing.T, branch string) string {
	t.Helper()
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}

	dir := t.TempDir()
	run := func(args ...string) {
		// Force-disable global git config that can break these tests on dev
		// machines (commit hooks set in ~/.gitconfig, gpg signing requirements,
		// commit signoff templates, etc.).
		base := []string{
			"-C", dir,
			"-c", "core.hooksPath=/dev/null",
			"-c", "commit.gpgSign=false",
			"-c", "tag.gpgSign=false",
		}
		cmd := exec.Command(gitBin, append(base, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	run("init", "--initial-branch="+branch)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("hi"), 0o644))
	run("add", "f")
	run("commit", "-m", "init")
	return dir
}

func TestResolveSymbolicHEAD_ReturnsFullRefOnBranch(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	dir := initTestRepoOnBranch(t, "feature/xyz")

	ref, err := resolveSymbolicHEAD(gitBin, dir)
	require.NoError(t, err)
	require.Equal(t, "refs/heads/feature/xyz", ref)
}

func TestResolveSymbolicHEAD_DetachedReturnsReferenceNotFound(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	dir := initTestRepoOnBranch(t, "master")

	// Detach HEAD by checking out the commit SHA directly.
	rev := exec.Command(gitBin, "-C", dir, "rev-parse", "HEAD")
	out, err := rev.Output()
	require.NoError(t, err)
	sha := string(out[:len(out)-1])

	checkout := exec.Command(gitBin, "-C", dir, "-c", "advice.detachedHead=false", "checkout", sha)
	checkoutOut, err := checkout.CombinedOutput()
	require.NoError(t, err, "checkout: %s", checkoutOut)

	_, err = resolveSymbolicHEAD(gitBin, dir)
	require.Error(t, err)
	require.ErrorIs(t, err, plumbing.ErrReferenceNotFound)
}
