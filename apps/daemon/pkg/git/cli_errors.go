// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daemon/pkg/childreap"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// gitCLIRun is the shared driver for clone, push, and pull on the git-CLI
// codepath. It looks up `git`, writes a one-shot GIT_ASKPASS helper, runs
// the supplied args with creds in env (never on argv / never in URL), and
// classifies the resulting error through wrapCLIError so the toolbox handler
// can map it to a sensible HTTP status.
//
// Params:
//   - op        human-readable label for error messages ("git clone", ...).
//   - args      git arguments — must NOT contain credentials.
//   - auth      basic auth; pass nil for local-only ops (askpass is still
//     written but never consulted — keeps the call sites symmetric).
//   - tailSize  bytes of stderr+stdout retained for the error message.
//
// Memory: the askpass dir is cleaned up via defer on every exit path,
// including panics. tailBuffer is bounded to tailSize regardless of how
// chatty git is.
func (s *Service) gitCLIRun(op string, args []string, auth *http.BasicAuth, tailSize int) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	askDir, err := os.MkdirTemp("", "daytona-git-*")
	if err != nil {
		return fmt.Errorf("create askpass temp dir: %w", err)
	}
	defer os.RemoveAll(askDir)

	askPath := filepath.Join(askDir, "askpass.sh")
	if err := os.WriteFile(askPath, []byte(askpassScript), 0o700); err != nil {
		return fmt.Errorf("write askpass helper: %w", err)
	}

	cmd := exec.Command(gitBin, args...)
	cmd.Env = buildCloneEnv(os.Environ(), askPath, auth)
	tail := s.attachCmdOutput(cmd, tailSize)

	// childreap.Run instead of cmd.Run so the reaper claiming the zombie
	// first doesn't surface as a spurious "wait: no child processes".
	exitCode, err := childreap.Run(cmd)
	if err != nil {
		return wrapCLIError(op, err, exitCode, tail.String(), auth)
	}
	if exitCode != 0 {
		return wrapCLIError(op, nil, exitCode, tail.String(), auth)
	}
	return nil
}

// classifyCLIError inspects captured `git` CLI stderr/stdout and returns the
// matching go-git sentinel error, or nil when no pattern matches. CLI paths
// wrap the returned sentinel via `%w` so the toolbox handler's
// `classifyGitError` (which uses `errors.Is` against go-git sentinels) maps
// CLI failures to the same HTTP status codes as the go-git default path
// (401/403/404/409) instead of an opaque 500.
//
// Patterns are matched against the lowercased output to be robust against
// minor casing differences between git versions.
func classifyCLIError(output string) error {
	s := strings.ToLower(output)

	switch {
	// 401 Unauthorized
	case contains(s, "authentication failed"),
		contains(s, "invalid username or password"),
		contains(s, "could not read username"),
		contains(s, "could not read password"),
		contains(s, "terminal prompts disabled"):
		return transport.ErrAuthenticationRequired

	// 403 Forbidden — distinct enough HTTP-403 markers; raw "forbidden"
	// alone is intentionally not matched because git surfaces it for both
	// auth-failures and authz-failures depending on the host.
	case contains(s, "the requested url returned error: 403"),
		contains(s, "remote: permission") && contains(s, "denied"):
		return transport.ErrAuthorizationFailed

	// 404 Not Found
	case contains(s, "repository not found"),
		contains(s, "the requested url returned error: 404"),
		contains(s, "remote: not found"),
		contains(s, "does not appear to be a git repository"):
		return transport.ErrRepositoryNotFound

	// 404 — local ref/branch resolution problems (e.g. detached HEAD push)
	case contains(s, "src refspec") && contains(s, "does not match any"),
		contains(s, "you are not currently on a branch"),
		contains(s, "unknown revision or path not in the working tree"):
		return plumbing.ErrReferenceNotFound

	// 409 Conflict — non-fast-forward / rejected updates.
	// Note: "failed to push some refs" is intentionally NOT matched — git
	// emits it for every push failure (including auth) and would shadow the
	// more-specific cases above.
	case contains(s, "non-fast-forward"),
		contains(s, "fetch first"),
		contains(s, "updates were rejected"),
		contains(s, "stale info"):
		return go_git.ErrNonFastForwardUpdate

	// 409 Conflict — pull with dirty worktree
	case contains(s, "your local changes to the following files would be overwritten"),
		contains(s, "please commit your changes or stash them"),
		contains(s, "needs merge"):
		return go_git.ErrWorktreeNotClean
	}

	return nil
}

// contains is a tiny wrapper so the switch cases above stay readable.
func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// redactCredentials replaces the basic-auth username and password with `***`
// in the given string. Used to scrub captured git output before it surfaces
// in error messages (where it would otherwise leak into logs / API responses).
//
// Order matters: replace the password first. If the username happens to be a
// substring of the password (e.g. username="foo", password="foopassword"),
// redacting the username first turns "foopassword" into "***password" and
// the remaining "password" tail leaks. Replacing the password first avoids
// this. We can't sort by length generically because the strings are short
// enough that an explicit ordering is clearer than a sort.
func redactCredentials(s string, auth *http.BasicAuth) string {
	if auth == nil {
		return s
	}
	if auth.Password != "" {
		s = strings.ReplaceAll(s, auth.Password, "***")
	}
	if auth.Username != "" {
		s = strings.ReplaceAll(s, auth.Username, "***")
	}
	return s
}

// wrapCLIError builds the final error returned from CLI codepaths. When
// `classifyCLIError` matches a known stderr pattern, the returned error wraps
// the corresponding go-git sentinel so the toolbox handler's
// `classifyGitError` can map it to the same HTTP status as the go-git path.
// Otherwise it falls back to the generic message (handled as 500 upstream).
//
// `op` is the human-readable operation name ("git push", "git clone", ...).
// `runErr` is non-nil when childreap.Run itself failed (couldn't start /
// reap the process). `exitCode` is 0 in that case.
//
// The captured git output is redacted of credentials before being included.
func wrapCLIError(op string, runErr error, exitCode int, output string, auth *http.BasicAuth) error {
	redacted := redactCredentials(output, auth)
	sentinel := classifyCLIError(output)

	var base error
	switch {
	case runErr != nil && sentinel != nil:
		base = fmt.Errorf("%s failed: %w\n--- git output (tail) ---\n%s\n--- cause: %v", op, sentinel, redacted, runErr)
	case runErr != nil:
		base = fmt.Errorf("%s failed: %w\n--- git output (tail) ---\n%s", op, runErr, redacted)
	case sentinel != nil:
		base = fmt.Errorf("%s exited with status %d: %w\n--- git output (tail) ---\n%s", op, exitCode, sentinel, redacted)
	default:
		base = fmt.Errorf("%s exited with status %d\n--- git output (tail) ---\n%s", op, exitCode, redacted)
	}
	return base
}
