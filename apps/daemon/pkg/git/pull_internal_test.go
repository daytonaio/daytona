// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"
)

var pullTestCreds = &http.BasicAuth{
	Username: "pull-test-user-xyz",
	Password: "pull-test-token-abc123",
}

func TestBuildPullArgs(t *testing.T) {
	got := buildPullArgs("/work-dir")
	require.Equal(t, []string{
		"-C", "/work-dir",
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"pull",
		"--ff-only",
		"--progress",
		"origin",
	}, got)
}

func TestBuildPullArgs_VerifiesTLS(t *testing.T) {
	// Pull must NOT skip TLS verification (parity with go-git PullOptions,
	// which does not set InsecureSkipTLS). Skipping verify would be a MITM
	// risk for the basic-auth token.
	args := buildPullArgs("/work-dir")
	for _, arg := range args {
		require.NotEqual(t, "http.sslVerify=false", arg,
			"pull args must NOT disable TLS verification")
	}
}

func TestBuildPullArgs_FastForwardOnly(t *testing.T) {
	// Pull must be fast-forward-only to match go-git's w.Pull() behavior
	// (which returns ErrNonFastForwardUpdate on divergent histories instead
	// of producing a merge commit).
	args := buildPullArgs("/work-dir")
	found := false
	for _, arg := range args {
		if arg == "--ff-only" {
			found = true
			break
		}
	}
	require.True(t, found, "pull args must include --ff-only")
}

func TestBuildPullArgs_NeverEmbedsCredsInArgs(t *testing.T) {
	args := buildPullArgs("/work-dir")

	for _, arg := range args {
		require.NotContains(t, arg, pullTestCreds.Username,
			"username leaked into pull args: %q", arg)
		require.NotContains(t, arg, pullTestCreds.Password,
			"password leaked into pull args: %q", arg)
	}
}

func TestBuildPullArgs_DisablesHooks(t *testing.T) {
	args := buildPullArgs("/work-dir")

	found := false
	for i, arg := range args {
		if arg == "core.hooksPath=/dev/null" && i > 0 && args[i-1] == "-c" {
			found = true
			break
		}
	}
	require.True(t, found, "pull args must disable hooks via core.hooksPath=/dev/null")
}

func TestPullCLI_EnvContainsAskpassAndCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", pullTestCreds)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
	require.Contains(t, env, "GIT_USERNAME="+pullTestCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+pullTestCreds.Password)
}

func TestPullCLI_EnvOmitsCredsWhenNil(t *testing.T) {
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
