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

func (s *Service) Push(auth *http.BasicAuth, remote, branch string, setUpstream bool) error {
	if isGitCLIModeEnabled() {
		return s.PushCLI(auth, remote, branch, setUpstream)
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	branchRef, err := resolvePushBranchRef(repo, branch)
	if err != nil {
		return err
	}

	remoteName := remote
	if remoteName == "" {
		remoteName = "origin"
	}

	options := &git.PushOptions{
		RemoteName: remoteName,
		Auth:       auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", branchRef, branchRef)),
		},
	}

	err = repo.Push(options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	if setUpstream {
		return s.setUpstreamConfig(repo, branchRef.Short(), remoteName, branchRef)
	}

	return nil
}

func resolvePushBranchRef(repo *git.Repository, branch string) (plumbing.ReferenceName, error) {
	if branch != "" {
		return plumbing.NewBranchReferenceName(branch), nil
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Name(), nil
}

// setUpstreamConfig records branch.<name>.{remote,merge} (git push --set-upstream).
func (s *Service) setUpstreamConfig(repo *git.Repository, branch, remote string, mergeRef plumbing.ReferenceName) error {
	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	if cfg.Branches == nil {
		cfg.Branches = map[string]*config.Branch{}
	}
	// Merge into any existing branch entry so we don't wipe other settings (e.g. rebase).
	b := cfg.Branches[branch]
	if b == nil {
		b = &config.Branch{}
	}
	b.Name = branch
	b.Remote = remote
	b.Merge = mergeRef
	cfg.Branches[branch] = b

	return repo.SetConfig(cfg)
}

func (s *Service) PushCLI(auth *http.BasicAuth, remote, branch string, setUpstream bool) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	branchRef := ""
	if branch != "" {
		branchRef = string(plumbing.NewBranchReferenceName(branch))
	} else {
		// Match go-git's `repo.Push` semantics, which pushes the resolved HEAD
		// ref as `<full-ref>:<full-ref>`. Resolve via the CLI rather than
		// re-opening the repo with go-git to keep this codepath dependency-free
		// (the whole point of the CLI fallback is bounded memory).
		branchRef, err = resolveSymbolicHEAD(gitBin, s.WorkDir)
		if err != nil {
			return err
		}
	}

	remoteName := remote
	if remoteName == "" {
		remoteName = "origin"
	}

	return s.runGitCLI(gitCLIOptions{
		op:       "git push",
		args:     buildPushArgs(s.WorkDir, remoteName, branchRef, setUpstream),
		auth:     auth,
		tailSize: 64 * 1024,
	})
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

func buildPushArgs(workDir, remote, branchRef string, setUpstream bool) []string {
	// Note: no -c http.sslVerify=false here. go-git's default PushOptions
	// does NOT skip TLS verification (unlike CloneOptions, where we set
	// InsecureSkipTLS:true). Skipping verify on push would be a behavior
	// change and a MITM risk for the basic-auth token.
	args := []string{
		"-C", workDir,
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"push",
	}
	if setUpstream {
		args = append(args, "--set-upstream")
	}
	args = append(args,
		remote,
		fmt.Sprintf("%s:%s", branchRef, branchRef),
		"--progress",
	)
	return args
}
