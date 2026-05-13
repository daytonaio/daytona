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
	"strconv"
	"strings"

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
	useCLI := os.Getenv(experimentalUseGitCloneCLIEnv) == "true"
	if err := validateCloneOptions(repo, useCLI); err != nil {
		return err
	}

	if useCLI {
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
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w\n--- git output (tail) ---\n%s", err, tail.String())
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		checkout := exec.Command(gitBin, buildCheckoutArgs(s.WorkDir, repo.Sha)...)
		// Checkout is a purely local op and does not need network credentials.
		// Pass a sanitized env with auth omitted so rogue checkout hooks cannot
		// exfiltrate GIT_USERNAME / GIT_PASSWORD.
		checkout.Env = buildCloneEnv(os.Environ(), askPath, nil)
		checkoutTail := s.attachCmdOutput(checkout, 16*1024)
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("git checkout %s failed: %w\n--- git output (tail) ---\n%s", repo.Sha, err, checkoutTail.String())
		}
	}

	if len(repo.SparsePaths) > 0 {
		sparseCheckout := exec.Command(gitBin, buildSparseCheckoutArgs(s.WorkDir, repo.SparsePaths)...)
		sparseCheckout.Env = buildCloneEnv(os.Environ(), askPath, nil)
		sparseCheckoutTail := s.attachCmdOutput(sparseCheckout, 16*1024)
		if err := sparseCheckout.Run(); err != nil {
			return fmt.Errorf("git sparse-checkout set failed: %w\n--- git output (tail) ---\n%s", err, sparseCheckoutTail.String())
		}
	}

	return nil
}

type InvalidCloneOptionsError struct {
	Message string
}

func (e *InvalidCloneOptionsError) Error() string {
	return e.Message
}

func validateCloneOptions(repo *gitprovider.GitRepository, useCLI bool) error {
	if repo == nil {
		return nil
	}

	if repo.Depth != nil && *repo.Depth < 1 {
		return &InvalidCloneOptionsError{Message: "depth must be greater than or equal to 1"}
	}
	if strings.ContainsAny(repo.Filter, "\x00\r\n") {
		return &InvalidCloneOptionsError{Message: "filter contains invalid characters"}
	}
	if strings.ContainsAny(repo.ReferencePath, "\x00\r\n") {
		return &InvalidCloneOptionsError{Message: "reference_path contains invalid characters"}
	}
	for _, sparsePath := range repo.SparsePaths {
		if err := validateSparsePath(sparsePath); err != nil {
			return err
		}
	}

	if hasOptimizedCloneOptions(repo) && !useCLI {
		return &InvalidCloneOptionsError{Message: "optimized clone options require DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI=true"}
	}
	if repo.Target == gitprovider.CloneTargetCommit && repo.Branch == "" && (repo.Depth != nil || repo.ShallowSince != "") {
		return &InvalidCloneOptionsError{Message: "commit_id with depth or shallow_since requires branch to be set"}
	}
	if repo.Dissociate != nil && *repo.Dissociate && repo.ReferencePath == "" {
		return &InvalidCloneOptionsError{Message: "dissociate requires reference_path"}
	}
	if repo.ShallowSubmodules != nil && *repo.ShallowSubmodules && !boolValue(repo.RecurseSubmodules) {
		return &InvalidCloneOptionsError{Message: "shallow_submodules requires recurse_submodules"}
	}
	if repo.FilterSubmodules != nil && *repo.FilterSubmodules {
		if !boolValue(repo.RecurseSubmodules) {
			return &InvalidCloneOptionsError{Message: "filter_submodules requires recurse_submodules"}
		}
		if repo.Filter == "" {
			return &InvalidCloneOptionsError{Message: "filter_submodules requires filter"}
		}
	}

	return nil
}

func hasOptimizedCloneOptions(repo *gitprovider.GitRepository) bool {
	return repo.Depth != nil ||
		repo.SingleBranch != nil ||
		repo.ShallowSince != "" ||
		repo.NoTags != nil ||
		repo.Filter != "" ||
		repo.Sparse != nil ||
		len(repo.SparsePaths) > 0 ||
		repo.ReferencePath != "" ||
		repo.Dissociate != nil ||
		repo.RecurseSubmodules != nil ||
		repo.ShallowSubmodules != nil ||
		repo.FilterSubmodules != nil
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func validateSparsePath(sparsePath string) error {
	if sparsePath == "" {
		return &InvalidCloneOptionsError{Message: "sparse_paths must not contain empty paths"}
	}
	if strings.ContainsAny(sparsePath, "\x00\r\n") {
		return &InvalidCloneOptionsError{Message: "sparse_paths contains invalid characters"}
	}
	if filepath.IsAbs(sparsePath) {
		return &InvalidCloneOptionsError{Message: "sparse_paths must contain relative paths"}
	}
	clean := filepath.Clean(sparsePath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return &InvalidCloneOptionsError{Message: "sparse_paths must contain relative paths"}
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
	}
	if repo.SingleBranch == nil || *repo.SingleBranch {
		args = append(args, "--single-branch")
	} else {
		args = append(args, "--no-single-branch")
	}
	args = append(args, "--progress")
	if repo.Depth != nil {
		args = append(args, "--depth="+strconv.Itoa(*repo.Depth))
	}
	if repo.ShallowSince != "" {
		args = append(args, "--shallow-since="+repo.ShallowSince)
	}
	if repo.NoTags != nil && *repo.NoTags {
		args = append(args, "--no-tags")
	}
	if repo.Filter != "" {
		args = append(args, "--filter="+repo.Filter)
	}
	if (repo.Sparse != nil && *repo.Sparse) || len(repo.SparsePaths) > 0 {
		args = append(args, "--sparse")
	}
	if repo.ReferencePath != "" {
		args = append(args, "--reference-if-able="+repo.ReferencePath)
	}
	if repo.Dissociate != nil && *repo.Dissociate {
		args = append(args, "--dissociate")
	}
	if repo.RecurseSubmodules != nil && *repo.RecurseSubmodules {
		args = append(args, "--recurse-submodules")
	}
	if repo.ShallowSubmodules != nil && *repo.ShallowSubmodules {
		args = append(args, "--shallow-submodules")
	}
	if repo.FilterSubmodules != nil && *repo.FilterSubmodules {
		args = append(args, "--also-filter-submodules")
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

func buildSparseCheckoutArgs(workDir string, sparsePaths []string) []string {
	args := []string{"-C", workDir, "sparse-checkout", "set", "--cone", "--"}
	args = append(args, sparsePaths...)
	return args
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
