// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git_test

import (
	"testing"

	"github.com/daytonaio/daemon/pkg/git"
	"github.com/stretchr/testify/suite"
)

type GitServiceTestSuite struct {
	suite.Suite
	gitService git.IGitService
}

func NewGitServiceTestSuite() *GitServiceTestSuite {
	return &GitServiceTestSuite{}
}

func (s *GitServiceTestSuite) SetupTest() {
	s.gitService = &git.Service{
		WorkDir: "/work-dir",
	}
}

func TestGitService(t *testing.T) {
	suite.Run(t, NewGitServiceTestSuite())
}
