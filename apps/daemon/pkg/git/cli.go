// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daemon/pkg/childreap"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type gitCLIOptions struct {
	op       string
	args     []string
	auth     *http.BasicAuth // when set, creds flow via a one-shot GIT_ASKPASS helper
	stdin    string
	redact   *http.BasicAuth // creds scrubbed from error output; defaults to auth
	tailSize int
}

// runGitCLI is the single entry point for git CLI invocations — the fallback for
// operations go-git cannot perform (clone/push/pull, reset --keep, credential approve).
func (s *Service) runGitCLI(opts gitCLIOptions) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	cmd := exec.Command(gitBin, opts.args...)

	if opts.auth != nil {
		askDir, err := os.MkdirTemp("", "daytona-git-*")
		if err != nil {
			return fmt.Errorf("create askpass temp dir: %w", err)
		}
		defer os.RemoveAll(askDir)

		askPath := filepath.Join(askDir, "askpass.sh")
		if err := os.WriteFile(askPath, []byte(askpassScript), 0o700); err != nil {
			return fmt.Errorf("write askpass helper: %w", err)
		}
		cmd.Env = buildGitCLIEnv(os.Environ(), askPath, opts.auth)
	} else {
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	}

	if opts.stdin != "" {
		cmd.Stdin = strings.NewReader(opts.stdin)
	}

	tail := s.attachCmdOutput(cmd, opts.tailSize)

	redact := opts.redact
	if redact == nil {
		redact = opts.auth
	}

	// childreap.Run (not cmd.Run) so a reaper claiming the zombie isn't surfaced as an error.
	exitCode, err := childreap.Run(cmd)
	if err != nil {
		return wrapCLIError(opts.op, err, exitCode, tail.String(), redact)
	}
	if exitCode != 0 {
		return wrapCLIError(opts.op, nil, exitCode, tail.String(), redact)
	}
	return nil
}

// GIT_ASKPASS helper: reads creds from env so they never hit argv, URL, or .git/config.
const askpassScript = `#!/bin/sh
case "$1" in
  Username*) printf '%s' "$GIT_USERNAME" ;;
  Password*) printf '%s' "$GIT_PASSWORD" ;;
esac
`

func buildGitCLIEnv(baseEnv []string, askPath string, auth *http.BasicAuth) []string {
	// glibc's getenv returns the first match, so appending can't override an
	// existing value. Strip conflicting keys from baseEnv before our values.
	managed := map[string]bool{
		"GIT_ASKPASS":         true,
		"GIT_TERMINAL_PROMPT": true,
		"GIT_USERNAME":        true,
		"GIT_PASSWORD":        true,
	}
	env := make([]string, 0, len(baseEnv)+4)
	for _, kv := range baseEnv {
		if i := strings.IndexByte(kv, '='); i > 0 && managed[kv[:i]] {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"GIT_ASKPASS="+askPath,
		"GIT_TERMINAL_PROMPT=0",
	)
	if auth != nil {
		env = append(env,
			"GIT_USERNAME="+auth.Username,
			"GIT_PASSWORD="+auth.Password,
		)
	}
	return env
}

// attachCmdOutput wires cmd.Stdout/Stderr to a bounded tail (returned so
// failures can include it) plus s.LogWriter when configured.
//
// Stdout and Stderr are assigned the same io.Writer value on purpose: per
// os/exec, when they're `==`-comparable and equal, at most one goroutine
// writes at a time — so the non-thread-safe tailBuffer / LogWriter stay safe
// without an explicit mutex.
func (s *Service) attachCmdOutput(cmd *exec.Cmd, tailSize int) *tailBuffer {
	tail := newTailBuffer(tailSize)
	var w io.Writer = tail
	if s.LogWriter != nil {
		w = io.MultiWriter(tail, s.LogWriter)
	}
	cmd.Stdout = w
	cmd.Stderr = w
	return tail
}

// tailBuffer keeps only the last N bytes — lets us include git's final error
// in wrapped errors without buffering gigabytes of progress output.
type tailBuffer struct {
	max int
	buf bytes.Buffer
}

func newTailBuffer(max int) *tailBuffer { return &tailBuffer{max: max} }

func (t *tailBuffer) Write(p []byte) (int, error) {
	n, err := t.buf.Write(p)
	if over := t.buf.Len() - t.max; over > 0 {
		t.buf.Next(over)
	}
	return n, err
}

func (t *tailBuffer) String() string { return t.buf.String() }
