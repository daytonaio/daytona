// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"strings"
	"testing"

	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"
)

func TestClassifyCLIError(t *testing.T) {
	cases := []struct {
		name     string
		output   string
		expected error
	}{
		// 401
		{
			name:     "authentication failed",
			output:   "remote: Invalid username or password.\nfatal: Authentication failed for 'https://github.com/foo/bar.git/'",
			expected: transport.ErrAuthenticationRequired,
		},
		{
			name:     "terminal prompts disabled",
			output:   "fatal: could not read Username for 'https://github.com': terminal prompts disabled",
			expected: transport.ErrAuthenticationRequired,
		},
		// 403
		{
			name:     "http 403",
			output:   "fatal: unable to access 'https://example.com/foo.git/': The requested URL returned error: 403",
			expected: transport.ErrAuthorizationFailed,
		},
		{
			name:     "permission denied",
			output:   "remote: Permission to foo/bar.git denied to user.",
			expected: transport.ErrAuthorizationFailed,
		},
		// 404 — remote
		{
			name:     "repository not found",
			output:   "remote: Repository not found.\nfatal: repository 'https://github.com/foo/missing.git/' not found",
			expected: transport.ErrRepositoryNotFound,
		},
		{
			name:     "http 404",
			output:   "fatal: unable to access 'https://example.com/foo.git/': The requested URL returned error: 404",
			expected: transport.ErrRepositoryNotFound,
		},
		// 404 — local ref (detached HEAD push, etc.)
		{
			name:     "src refspec does not match",
			output:   "error: src refspec HEAD does not match any.\nerror: failed to push some refs to 'origin'",
			expected: plumbing.ErrReferenceNotFound,
		},
		// 409 — non-fast-forward
		{
			name:     "non-fast-forward",
			output:   "! [rejected]        main -> main (non-fast-forward)\nerror: failed to push some refs to 'origin'",
			expected: go_git.ErrNonFastForwardUpdate,
		},
		{
			name:     "fetch first",
			output:   "! [rejected]        main -> main (fetch first)\nerror: failed to push some refs to 'origin'\nhint: Updates were rejected because the remote contains work that you do not have locally.",
			expected: go_git.ErrNonFastForwardUpdate,
		},
		// 409 — pull onto dirty worktree
		{
			name:     "local changes would be overwritten",
			output:   "error: Your local changes to the following files would be overwritten by merge:\n\tREADME.md\nPlease commit your changes or stash them before you merge.",
			expected: go_git.ErrWorktreeNotClean,
		},
		// No match
		{
			name:     "unknown error",
			output:   "fatal: something completely unexpected happened",
			expected: nil,
		},
		{
			name:     "empty output",
			output:   "",
			expected: nil,
		},
		// Auth-failure shadow check: "failed to push some refs" alone must
		// not be classified as non-fast-forward.
		{
			name:     "generic push failure is not classified as conflict",
			output:   "error: failed to push some refs to 'origin'",
			expected: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := classifyCLIError(tc.output)
			if tc.expected == nil {
				require.Nil(t, got, "expected no match, got %v", got)
				return
			}
			require.ErrorIs(t, got, tc.expected)
		})
	}
}

func TestWrapCLIError_AttachesSentinelForKnownPatterns(t *testing.T) {
	auth := &http.BasicAuth{Username: "u", Password: "secret-token"}

	// runErr nil, exitCode != 0, output matches auth-failure pattern → wraps sentinel
	err := wrapCLIError("git push", nil, 128,
		"fatal: Authentication failed for 'https://github.com/foo/bar.git/'", auth)
	require.Error(t, err)
	require.ErrorIs(t, err, transport.ErrAuthenticationRequired)
}

func TestWrapCLIError_FallsBackWhenUnknownPattern(t *testing.T) {
	auth := &http.BasicAuth{Username: "u", Password: "secret-token"}

	err := wrapCLIError("git push", nil, 1, "fatal: kaboom", auth)
	require.Error(t, err)
	// Should NOT match any go-git sentinel (will map to 500 upstream).
	require.False(t, errors.Is(err, transport.ErrAuthenticationRequired))
	require.False(t, errors.Is(err, transport.ErrRepositoryNotFound))
	require.False(t, errors.Is(err, go_git.ErrNonFastForwardUpdate))
}

func TestWrapCLIError_RedactsCredentialsInOutput(t *testing.T) {
	auth := &http.BasicAuth{Username: "leaky-user", Password: "leaky-token-zzz"}

	output := "fatal: Authentication failed for 'https://leaky-user:leaky-token-zzz@example.com/foo.git/'"
	err := wrapCLIError("git push", nil, 128, output, auth)
	require.Error(t, err)

	msg := err.Error()
	require.NotContains(t, msg, "leaky-user")
	require.NotContains(t, msg, "leaky-token-zzz")
	require.Contains(t, msg, "***")
}

func TestWrapCLIError_IncludesRunErrCause(t *testing.T) {
	auth := &http.BasicAuth{Username: "u", Password: "p"}

	runErr := errors.New("exec: signal killed")
	err := wrapCLIError("git pull", runErr, 0, "(no output)", auth)
	require.Error(t, err)
	// runErr should be surfaced (either as the wrapped error or in the message).
	require.True(t,
		errors.Is(err, runErr) || strings.Contains(err.Error(), runErr.Error()),
		"expected runErr to be referenced; got: %v", err)
}

func TestRedactCredentials(t *testing.T) {
	auth := &http.BasicAuth{
		Username: "secret-user",
		Password: "secret-token-xyz",
	}

	t.Run("redacts username and password", func(t *testing.T) {
		raw := "fatal: Authentication failed for 'https://secret-user@github.com/repo.git'\nPassword was secret-token-xyz"
		got := redactCredentials(raw, auth)
		require.NotContains(t, got, "secret-user")
		require.NotContains(t, got, "secret-token-xyz")
		require.Contains(t, got, "***")
	})

	t.Run("nil auth returns input unchanged", func(t *testing.T) {
		raw := "some output with no secrets"
		got := redactCredentials(raw, nil)
		require.Equal(t, raw, got)
	})

	t.Run("empty creds are no-ops", func(t *testing.T) {
		raw := "some output"
		got := redactCredentials(raw, &http.BasicAuth{})
		require.Equal(t, raw, got)
	})

	t.Run("username substring of password does not leak password tail", func(t *testing.T) {
		// Regression: when username is a substring of password (e.g.
		// username="foo", password="foopassword"), redacting the username
		// first would turn "foopassword" into "***password" and leak the
		// "password" suffix. Password must be redacted first.
		overlapping := &http.BasicAuth{
			Username: "foo",
			Password: "foopassword",
		}
		raw := "leaked: foopassword and foo"
		got := redactCredentials(raw, overlapping)
		require.NotContains(t, got, "password", "password tail leaked: %q", got)
		require.NotContains(t, got, "foopassword")
	})
}
