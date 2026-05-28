// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Pull(auth *http.BasicAuth) error {
	if isGitCLIModeEnabled() {
		return s.PullCLI(auth)
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	options := &git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	}

	return w.Pull(options)
}

func (s *Service) PullCLI(auth *http.BasicAuth) error {
	return s.gitCLIRun("git pull", buildPullArgs(s.WorkDir), auth, 64*1024)
}

func buildPullArgs(workDir string) []string {
	// --ff-only matches go-git's `w.Pull()` semantics, which fails with
	// ErrNonFastForwardUpdate on divergent histories instead of producing
	// a merge commit. Without --ff-only, plain `git pull` would honor the
	// repo/user pull.rebase / merge.ff config and could diverge from the
	// go-git path's behavior.
	return []string{
		"-C", workDir,
		"-c", "credential.helper=",
		"-c", "core.hooksPath=/dev/null",
		"pull",
		"--ff-only",
		"--progress",
		"origin",
	}
}
