// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"slices"
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

	// Without auth, credential env vars must not leak.
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return len(s) > len("GIT_USERNAME=") && s[:len("GIT_USERNAME=")] == "GIT_USERNAME="
	}), "GIT_USERNAME must not be set when auth is nil")
	require.False(t, slices.ContainsFunc(env, func(s string) bool {
		return len(s) > len("GIT_PASSWORD=") && s[:len("GIT_PASSWORD=")] == "GIT_PASSWORD="
	}), "GIT_PASSWORD must not be set when auth is nil")
}

func TestBuildCloneEnv_PreservesBaseEnv(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/daytona"}
	env := buildCloneEnv(base, "/tmp/askpass.sh", nil)

	require.Contains(t, env, "PATH=/usr/bin")
	require.Contains(t, env, "HOME=/home/daytona")
}

func TestBuildCheckoutArgs(t *testing.T) {
	got := buildCheckoutArgs("/work-dir", "1234567890")
	require.Equal(t, []string{"-C", "/work-dir", "checkout", "--", "1234567890"}, got)
}
