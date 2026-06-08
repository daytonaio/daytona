// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Push(auth *http.BasicAuth) error {
	if isGitCLIModeEnabled() {
		return s.PushCLI(auth)
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	options := &git.PushOptions{
		Auth: auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", ref.Name(), ref.Name())),
		},
	}

	return repo.Push(options)
}

func (s *Service) PushCLI(auth *http.BasicAuth) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	// Match go-git's `repo.Push` semantics, which pushes the resolved HEAD
	// ref as `<full-ref>:<full-ref>`. Resolve via the CLI rather than
	// re-opening the repo with go-git to keep this codepath dependency-free
	// (the whole point of the CLI fallback is bounded memory).
	branchRef, err := resolveSymbolicHEAD(gitBin, s.WorkDir)
	if err != nil {
		return err
	}

	return s.gitCLIRun("git push", buildPushArgs(s.WorkDir, branchRef), auth, 64*1024)
}

// resolveSymbolicHEAD returns the fully-qualified ref name of the current
// branch (e.g. `refs/heads/main`). When HEAD is detached, it returns a
// plumbing.ErrReferenceNotFound-wrapped error so the toolbox handler maps it
// to HTTP 404 — same status the go-git path produces for an unresolvable ref.
func resolveSymbolicHEAD(gitBin, workDir string) (string, error) {
	cmd := exec.Command(gitBin, "-C", workDir, "symbolic-ref", "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// `git symbolic-ref HEAD` exits non-zero in detached HEAD state. We
		// can't push HEAD on detached state — go-git fails too in that case
		// (different error, same outcome). Surface a 404-mappable error.
		return "", fmt.Errorf("resolve HEAD: detached or unresolvable (%s): %w",
			strings.TrimSpace(stderr.String()), plumbing.ErrReferenceNotFound)
	}
	ref := strings.TrimSpace(stdout.String())
	if ref == "" {
		return "", fmt.Errorf("resolve HEAD: empty symbolic ref: %w", plumbing.ErrReferenceNotFound)
	}
	return ref, nil
}

func buildPushArgs(workDir, branchRef string) []string {
	return []string{
		"-C", workDir,
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"push",
		"origin",
		fmt.Sprintf("%s:%s", branchRef, branchRef),
		"--progress",
	}
}
