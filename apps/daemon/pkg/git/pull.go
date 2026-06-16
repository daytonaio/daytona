// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Pull(auth *http.BasicAuth, remote, branch string) error {
	if isGitCLIModeEnabled() {
		return s.PullCLI(auth, remote, branch)
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	remoteName := remote
	if remoteName == "" {
		remoteName = "origin"
	}

	options := &git.PullOptions{
		RemoteName: remoteName,
		Auth:       auth,
	}
	if branch != "" {
		options.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	return w.Pull(options)
}

func (s *Service) PullCLI(auth *http.BasicAuth, remote, branch string) error {
	return s.runGitCLI(gitCLIOptions{
		op:       "git pull",
		args:     buildPullArgs(s.WorkDir, remote, branch),
		auth:     auth,
		tailSize: 64 * 1024,
	})
}

func buildPullArgs(workDir, remote, branch string) []string {
	remoteName := remote
	if remoteName == "" {
		remoteName = "origin"
	}

	// --ff-only matches go-git's `w.Pull()` semantics, which fails with
	// ErrNonFastForwardUpdate on divergent histories instead of producing
	// a merge commit. Without --ff-only, plain `git pull` would honor the
	// repo/user pull.rebase / merge.ff config and could diverge from the
	// go-git path's behavior.
	args := []string{
		"-C", workDir,
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"pull",
		"--ff-only",
		"--progress",
		remoteName,
	}
	if branch != "" {
		args = append(args, branch)
	}
	return args
}
