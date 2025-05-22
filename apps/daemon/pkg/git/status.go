// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
)

func (s *Service) GetGitStatus() (*GitStatus, error) {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	files := []*FileStatus{}
	for path, file := range status {
		files = append(files, &FileStatus{
			Name:     path,
			Extra:    file.Extra,
			Staging:  MapStatus[file.Staging],
			Worktree: MapStatus[file.Worktree],
		})
	}

	branchPublished, err := s.isBranchPublished()
	if err != nil {
		return nil, err
	}

	ahead, behind, err := s.getAheadBehindInfo()
	if err != nil {
		return nil, err
	}

	return &GitStatus{
		CurrentBranch:   ref.Name().Short(),
		Files:           files,
		BranchPublished: branchPublished,
		Ahead:           ahead,
		Behind:          behind,
	}, nil
}

func (s *Service) isBranchPublished() (bool, error) {
	upstream, err := s.getUpstreamBranch()
	if err != nil {
		return false, err
	}
	return upstream != "", nil
}

func (s *Service) getUpstreamBranch() (string, error) {
	cmd := exec.Command("git", "-C", s.ProjectDir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(string(out)), nil
}

func (s *Service) getAheadBehindInfo() (int, int, error) {
	upstream, err := s.getUpstreamBranch()
	if err != nil {
		return 0, 0, err
	}
	if upstream == "" {
		return 0, 0, nil
	}

	cmd := exec.Command("git", "-C", s.ProjectDir, "rev-list", "--left-right", "--count", fmt.Sprintf("%s...HEAD", upstream))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, nil
	}

	return parseAheadBehind(out)
}

func parseAheadBehind(output []byte) (int, int, error) {
	counts := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(counts) != 2 {
		return 0, 0, nil
	}

	ahead, err := strconv.Atoi(counts[1])
	if err != nil {
		return 0, 0, nil
	}

	behind, err := strconv.Atoi(counts[0])
	if err != nil {
		return 0, 0, nil
	}

	return ahead, behind, nil
}
