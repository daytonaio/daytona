// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"slices"
	"strings"
	"testing"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"
)

// Deliberately distinct from anything that could appear in test URLs, so
// credential-leak regression checks don't false-positive on URL components.
var testCreds = &http.BasicAuth{
	Username: "test-user-xyz",
	Password: "test-token-abc123",
}

func TestBuildCloneArgs(t *testing.T) {
	tests := []struct {
		name     string
		repo     *gitprovider.GitRepository
		workDir  string
		expected []string
	}{
		{
			name: "https URL with branch",
			repo: &gitprovider.GitRepository{
				Url:    "https://github.com/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "http URL with branch",
			repo: &gitprovider.GitRepository{
				Url:    "http://localhost:3000/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "http://localhost:3000/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "URL without protocol gets https:// prefix",
			repo: &gitprovider.GitRepository{
				Url:    "github.com/daytonaio/daytona",
				Branch: "main",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--branch", "main",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
		{
			name: "no branch omits --branch flag",
			repo: &gitprovider.GitRepository{
				Url: "https://github.com/daytonaio/daytona",
			},
			workDir: "/work-dir",
			expected: []string{
				"-c", "credential.helper=",
				"-c", "http.sslVerify=false",
				"clone", "--single-branch", "--progress",
				"--", "https://github.com/daytonaio/daytona", "/work-dir",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCloneArgs(tc.repo, tc.workDir)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildCloneArgs_NeverEmbedsCredsInURL(t *testing.T) {
	// Regression guard: the old []string-returning implementation embedded
	// user:pass@ in the URL, which persisted into .git/config. The new impl
	// must never do that — creds flow through GIT_ASKPASS env only.
	repo := &gitprovider.GitRepository{
		Url:    "https://github.com/daytonaio/daytona",
		Branch: "main",
	}
	args := buildCloneArgs(repo, "/work-dir")

	for _, arg := range args {
		require.NotContains(t, arg, testCreds.Username, "username leaked into clone args: %q", arg)
		require.NotContains(t, arg, testCreds.Password, "password leaked into clone args: %q", arg)
		require.NotContains(t, arg, "@github.com", "credential-in-URL pattern leaked into clone args: %q", arg)
	}
}

func TestBuildCloneEnv_WithCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", testCreds)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
	require.Contains(t, env, "GIT_USERNAME="+testCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+testCreds.Password)
}

func TestBuildCloneEnv_WithoutCreds(t *testing.T) {
	env := buildCloneEnv(nil, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_TERMINAL_PROMPT=0")

	// Without auth, credential env vars must not leak — even as empty values.
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return strings.HasPrefix(s, "GIT_USERNAME=")
	}), "GIT_USERNAME must not be set when auth is nil")
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return strings.HasPrefix(s, "GIT_PASSWORD=")
	}), "GIT_PASSWORD must not be set when auth is nil")
}

// TestBuildCloneEnv_OverridesBaseEnv covers the glibc getenv-first-match
// quirk: if baseEnv already has any of our managed keys, buildCloneEnv must
// strip them so our values take effect inside the subprocess.
func TestBuildCloneEnv_OverridesBaseEnv(t *testing.T) {
	base := []string{
		"PATH=/usr/bin",
		"GIT_ASKPASS=/wrong/path",
		"GIT_USERNAME=inherited-user",
		"GIT_PASSWORD=inherited-pass",
		"HOME=/home/daytona",
	}
	env := buildCloneEnv(base, "/tmp/askpass.sh", testCreds)

	// Base env unrelated to git is preserved.
	require.Contains(t, env, "PATH=/usr/bin")
	require.Contains(t, env, "HOME=/home/daytona")

	// Managed keys from baseEnv are dropped; our values are the only ones.
	countPrefix := func(prefix string) int {
		n := 0
		for _, kv := range env {
			if strings.HasPrefix(kv, prefix) {
				n++
			}
		}
		return n
	}
	require.Equal(t, 1, countPrefix("GIT_ASKPASS="))
	require.Equal(t, 1, countPrefix("GIT_USERNAME="))
	require.Equal(t, 1, countPrefix("GIT_PASSWORD="))

	require.Contains(t, env, "GIT_ASKPASS=/tmp/askpass.sh")
	require.Contains(t, env, "GIT_USERNAME="+testCreds.Username)
	require.Contains(t, env, "GIT_PASSWORD="+testCreds.Password)
}

func TestBuildCloneEnv_PreservesBaseEnv(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/daytona"}
	env := buildCloneEnv(base, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "PATH=/usr/bin")
	require.Contains(t, env, "HOME=/home/daytona")
}

func TestBuildCheckoutArgs(t *testing.T) {
	got := buildCheckoutArgs("/work-dir", "1234567890")
	// Must NOT include "--" before the SHA — git would then treat it as a
	// pathspec and fail with "did not match any file(s) known to git".
	require.Equal(t, []string{"-C", "/work-dir", "checkout", "1234567890"}, got)
	require.NotContains(t, got, "--", "checkout args must not include -- before the SHA")
}
