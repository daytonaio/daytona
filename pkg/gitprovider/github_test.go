// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GitHubGitProviderTestSuite struct {
	gitProvider *GitHubGitProvider
	suite.Suite
}

func NewGitHubGitProviderTestSuite() *GitHubGitProviderTestSuite {
	return &GitHubGitProviderTestSuite{
		gitProvider: NewGitHubGitProvider("", nil),
	}
}

func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://github.com/daytonaio/daytona/pull/1"
	prContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://github.com/daytonaio/daytona.git",
		Source:   "github.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: &[]uint32{1}[0],
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(httpContext, prContext)
}

func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://github.com/daytonaio/daytona/blob/main/README.md"
	blobContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Source:   "github.com",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   &[]string{"main"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     &[]string{"README.md"}[0],
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://github.com/daytonaio/daytona/tree/test-branch"
	branchContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Source:   "github.com",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   &[]string{"test-branch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(httpContext, branchContext)
}

func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_BranchNameWithSlash() {
	commitUrl := "https://github.com/daytonaio/daytona/tree/test/test-branch"
	commitContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Source:   "github.com",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   &[]string{"test/test-branch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}
func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://github.com/daytonaio/daytona/commits/test-branch"
	commitsContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Source:   "github.com",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   &[]string{"test-branch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (g *GitHubGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://github.com/daytonaio/daytona/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Source:   "github.com",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   &[]string{"COMMIT_SHA"}[0],
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *GitHubGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "github.com",
		Url:    "https://github.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona", url)
}

func (g *GitHubGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "github.com",
		Url:    "https://github.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona/tree/test-branch", url)
}

func (g *GitHubGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "github.com",
		Url:    "https://github.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona/tree/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona/blob/main/README.md", url)
}

func (g *GitHubGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "github.com",
		Url:    "https://github.com/daytonaio/daytona.git",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://github.com/daytonaio/daytona/commit/COMMIT_SHA", url)
}

func TestGitHubGitProvider(t *testing.T) {
	suite.Run(t, NewGitHubGitProviderTestSuite())
}
