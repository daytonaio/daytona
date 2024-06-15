// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GitNessGitProviderTestSuite struct {
	gitProvider *GitNessGitProvider
	suite.Suite
}

func NewGitNessGitProviderTestSuite() *GitNessGitProviderTestSuite {
	baseApiUrl := "http://localhost:3000/api/v1/"
	token := ""
	return &GitNessGitProviderTestSuite{
		gitProvider: NewGitNessGitProvider(token, &baseApiUrl),
	}
}

func (g *GitNessGitProviderTestSuite) TestGetNamespaces() {
	namespaces, err := g.gitProvider.GetNamespaces()

	require := g.Require()
	require.Nil(err)
	require.NotEmpty(namespaces)
}

func (g *GitNessGitProviderTestSuite) TestGetRepositories() {
	repositories, err := g.gitProvider.GetRepositories("test")

	require := g.Require()
	require.Nil(err)
	require.NotEmpty(repositories)
}

func (g *GitNessGitProviderTestSuite) TestGetRepoBranches() {
	branches, err := g.gitProvider.GetRepoBranches("test", "test")

	require := g.Require()
	require.Nil(err)
	require.NotEmpty(branches)
}

func (g *GitNessGitProviderTestSuite) TestGetRepoPRs() {
	prs, err := g.gitProvider.GetRepoPRs("test", "test")

	require := g.Require()
	require.Nil(err)
	require.NotEmpty(prs)
}

func (g *GitNessGitProviderTestSuite) TestGetUser() {
	user, err := g.gitProvider.GetUser()

	require := g.Require()
	require.Nil(err)
	require.NotNil(user)
}

func (g *GitNessGitProviderTestSuite) TestGetLastCommitSha() {
	commitSha, err := g.gitProvider.GetLastCommitSha(&StaticGitContext{
		Url:    "https://localhost:3000/test/test",
		Branch: &[]string{"main"}[0],
	})

	require := g.Require()
	require.Nil(err)
	require.NotEmpty(commitSha)
}

func (g *GitNessGitProviderTestSuite) TestParseStaticGitContext_PR() {
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

	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *GitNessGitProviderTestSuite) TestParseStaticGitContext_Files() {
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

	httpContext, err := g.gitProvider.parseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(blobContext, httpContext)
}

func (g *GitNessGitProviderTestSuite) TestParseStaticGitContext_Branch() {
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

	httpContext, err := g.gitProvider.parseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(branchContext, httpContext)
}

func (g *GitNessGitProviderTestSuite) TestParseStaticGitContext_Commits() {
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

	httpContext, err := g.gitProvider.parseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func (g *GitNessGitProviderTestSuite) TestParseStaticGitContext_Commit() {
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

	httpContext, err := g.gitProvider.parseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func TestGitNessGitProvider(t *testing.T) {
	suite.Run(t, NewGitNessGitProviderTestSuite())
}
