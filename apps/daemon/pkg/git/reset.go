// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Reset resets HEAD to target (HEAD when empty). go-git lacks "keep", which
// falls back to the git CLI.
func (s *Service) Reset(mode, target string, files []string) error {
	if mode == "keep" {
		return s.resetKeepCLI(target, files)
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	opts := &git.ResetOptions{
		Files: files,
	}

	switch mode {
	case "", "mixed":
		opts.Mode = git.MixedReset
	case "soft":
		opts.Mode = git.SoftReset
	case "hard":
		opts.Mode = git.HardReset
	case "merge":
		opts.Mode = git.MergeReset
	default:
		return fmt.Errorf("unsupported reset mode %q (supported: soft, mixed, hard, merge, keep)", mode)
	}

	if target != "" {
		hash, err := repo.ResolveRevision(plumbing.Revision(target))
		if err != nil {
			return err
		}
		opts.Commit = *hash
	}

	return worktree.Reset(opts)
}

func (s *Service) resetKeepCLI(target string, files []string) error {
	args := []string{
		"-C", s.WorkDir,
		"-c", "core.hooksPath=/dev/null",
		"reset", "--keep",
	}
	if target != "" {
		args = append(args, target)
	}
	if len(files) > 0 {
		args = append(args, "--")
		args = append(args, files...)
	}

	return s.runGitCLI(gitCLIOptions{op: "git reset", args: args, tailSize: 16 * 1024})
}
