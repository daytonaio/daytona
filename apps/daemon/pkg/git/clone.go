// Copyright 2025 Daytona Platforms Inc.
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
	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Set to "true" to opt into the git-CLI clone path (bounded memory, needs `git` in PATH).
const experimentalUseGitCloneCLIEnv = "DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI"

func (s *Service) CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error {
	if os.Getenv(experimentalUseGitCloneCLIEnv) == "true" {
		return s.CloneRepositoryCLI(repo, auth)
	}

	cloneOptions := &git.CloneOptions{
		URL:             repo.Url,
		SingleBranch:    true,
		InsecureSkipTLS: true,
		Auth:            auth,
	}

	if s.LogWriter != nil {
		cloneOptions.Progress = s.LogWriter
	}

	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// This can be removed once go-git implements the git v2 protocol.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	if repo.Branch != "" {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(repo.Branch)
	}

	_, err := git.PlainClone(s.WorkDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		r, err := git.PlainOpen(s.WorkDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(repo.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

// GIT_ASKPASS helper: reads creds from env so they never hit argv, URL, or .git/config.
const askpassScript = `#!/bin/sh
case "$1" in
  Username*) printf '%s' "$GIT_USERNAME" ;;
  Password*) printf '%s' "$GIT_PASSWORD" ;;
esac
`

// CloneRepositoryCLI clones via the `git` CLI. Bounded memory (mmap pack handling).
// Creds flow through GIT_ASKPASS + env — never via URL or argv.
func (s *Service) CloneRepositoryCLI(repo *gitprovider.GitRepository, auth *http.BasicAuth) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	askDir, err := os.MkdirTemp("", "daytona-clone-*")
	if err != nil {
		return fmt.Errorf("create askpass temp dir: %w", err)
	}
	defer os.RemoveAll(askDir)

	askPath := filepath.Join(askDir, "askpass.sh")
	if err := os.WriteFile(askPath, []byte(askpassScript), 0o700); err != nil {
		return fmt.Errorf("write askpass helper: %w", err)
	}

	cmd := exec.Command(gitBin, buildCloneArgs(repo, s.WorkDir)...)
	cmd.Env = buildCloneEnv(os.Environ(), askPath, auth)
	tail := s.attachCmdOutput(cmd, 64*1024)
	// childreap.Run instead of cmd.Run so the reaper claiming the zombie
	// first doesn't surface as a spurious "git clone failed: wait: no
	// child processes" to API clients.
	exitCode, err := childreap.Run(cmd)
	if err != nil {
		return fmt.Errorf("git clone failed: %w\n--- git output (tail) ---\n%s", err, tail.String())
	}
	if exitCode != 0 {
		return fmt.Errorf("git clone exited with status %d\n--- git output (tail) ---\n%s", exitCode, tail.String())
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		checkout := exec.Command(gitBin, buildCheckoutArgs(s.WorkDir, repo.Sha)...)
		// Checkout is a purely local op and does not need network credentials.
		// Pass a sanitized env with auth omitted so rogue checkout hooks cannot
		// exfiltrate GIT_USERNAME / GIT_PASSWORD.
		checkout.Env = buildCloneEnv(os.Environ(), askPath, nil)
		checkoutTail := s.attachCmdOutput(checkout, 16*1024)
		checkoutCode, err := childreap.Run(checkout)
		if err != nil {
			return fmt.Errorf("git checkout %s failed: %w\n--- git output (tail) ---\n%s", repo.Sha, err, checkoutTail.String())
		}
		if checkoutCode != 0 {
			return fmt.Errorf("git checkout %s exited with status %d\n--- git output (tail) ---\n%s", repo.Sha, checkoutCode, checkoutTail.String())
		}
	}

	return nil
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

// Credentials must NEVER be embedded in the URL — they flow via GIT_ASKPASS (see buildCloneEnv).
func buildCloneArgs(repo *gitprovider.GitRepository, workDir string) []string {
	cloneURL := repo.Url
	if !strings.Contains(cloneURL, "://") {
		cloneURL = "https://" + cloneURL
	}

	args := []string{
		"-c", "credential.helper=", // prevent any inherited helper from persisting the token
		"-c", "http.sslVerify=false", // parity with go-git InsecureSkipTLS
		"clone",
		"--single-branch",
		"--progress",
	}
	if repo.Branch != "" {
		args = append(args, "--branch", repo.Branch)
	}
	args = append(args, "--", cloneURL, workDir)
	return args
}

func buildCloneEnv(baseEnv []string, askPath string, auth *http.BasicAuth) []string {
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

func buildCheckoutArgs(workDir, sha string) []string {
	// No `--` separator: that would make git treat the SHA as a pathspec.
	return []string{"-C", workDir, "checkout", sha}
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
