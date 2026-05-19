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
	return []string{
		"-C", workDir,
		"-c", "credential.helper=",
		"-c", "http.sslVerify=false",
		"-c", "core.hooksPath=/dev/null",
		"pull",
		"origin",
		"--progress",
	}
}
