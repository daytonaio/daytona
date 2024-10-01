// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type GitnessGitProviderTestSuite struct {
	gitProvider *GitnessGitProvider
	suite.Suite
}

func NewGitnessGitProviderTestSuite() *GitnessGitProviderTestSuite {
	return &GitnessGitProviderTestSuite{
		gitProvider: NewGitnessGitProvider("", "http://localhost:3000"),
	}
}

func (g *GitnessGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://localhost:3000/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *GitnessGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *GitnessGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://localhost:3000/test/test/pulls/1"
	prContext := &StaticGitContext{
		Id:       "test",
		Name:     "test",
		Owner:    "test",
		Url:      "https://localhost:3000/git/test/test.git",
		Branch:   nil,
		Sha:      nil,
		Source:   "localhost:3000",
		Path:     nil,
		PrNumber: &[]uint32{1}[0],
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *GitnessGitProviderTestSuite) TestParseStaticGitContext_Files() {
	blobUrl := "https://localhost:3000/test/test/files/main/~/test.md"
	blobContext := &StaticGitContext{
		Id:       "test",
		Name:     "test",
		Owner:    "test",
		Source:   "localhost:3000",
		Url:      "https://localhost:3000/git/test/test.git",
		Branch:   &[]string{"main"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     &[]string{"~/test.md"}[0],
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)
	require.Nil(err)
	require.Equal(blobContext, httpContext)
}

func (g *GitnessGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://localhost:3000/test/test/files/newbranch"
	branchContext := &StaticGitContext{
		Id:       "test",
		Name:     "test",
		Owner:    "test",
		Source:   "localhost:3000",
		Url:      "https://localhost:3000/git/test/test.git",
		Branch:   &[]string{"newbranch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(branchUrl)
	require.Nil(err)
	require.Equal(branchContext, httpContext)
}

func (g *GitnessGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitUrl := "https://localhost:3000/test/test/commits/newbranch"
	commitContext := &StaticGitContext{
		Id:       "test",
		Name:     "test",
		Owner:    "test",
		Source:   "localhost:3000",
		Url:      "https://localhost:3000/git/test/test.git",
		Branch:   &[]string{"newbranch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)
	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func (g *GitnessGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://localhost:3000/test/test/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "test",
		Name:     "test",
		Owner:    "test",
		Source:   "localhost:3000",
		Url:      "https://localhost:3000/git/test/test.git",
		Branch:   &[]string{"COMMIT_SHA"}[0],
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)
	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func (g *GitnessGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("localhost:3000"),
		Url:    "https://localhost:3000/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona", url)
}

func (g *GitnessGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("localhost:3000"),
		Url:    "https://localhost:3000/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona/files/test-branch", url)
}

func (g *GitnessGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("localhost:3000"),
		Url:    "https://localhost:3000/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona/files/test-branch/~/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona/files/main/~/README.md", url)
}

func (g *GitnessGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("localhost:3000"),
		Url:    "https://localhost:3000/daytonaio/daytona.git",
		Branch: util.Pointer("COMMIT_SHA"),
		Sha:    util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona/files/COMMIT_SHA/~/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://localhost:3000/daytonaio/daytona/files/COMMIT_SHA", url)
}

func TestGitnessGitProvider(t *testing.T) {
	suite.Run(t, NewGitnessGitProviderTestSuite())
}
