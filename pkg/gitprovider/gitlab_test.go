// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GitLabGitProviderTestSuite struct {
	gitProvider *GitLabGitProvider
	suite.Suite
}

func NewGitLabGitProviderTestSuite() *GitLabGitProviderTestSuite {
	return &GitLabGitProviderTestSuite{
		gitProvider: NewGitLabGitProvider("", nil),
	}
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_HTTP() {
	httpSimple := "https://gitlab.com/gitlab-org/gitlab"
	simpleContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(httpSimple)

	require.Nil(err)
	require.Equal(httpContext, simpleContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_HTTP_Subgroups() {
	httpSimple := "https://gitlab.com/gitlab-org/subgroup1/subgroup2/gitlab"
	simpleContext := &StaticGitContext{
		Id:       "gitlab-org/subgroup1/subgroup2/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org/subgroup1/subgroup2",
		Url:      "https://gitlab.com/gitlab-org/subgroup1/subgroup2/gitlab.git",
		Source:   "gitlab.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(httpSimple)

	require.Nil(err)
	require.Equal(httpContext, simpleContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_MR() {
	mrUrl := "https://gitlab.com/gitlab-org/gitlab/-/merge_requests/1"
	mrContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: &[]uint32{1}[0],
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(mrUrl)

	require.Nil(err)
	require.Equal(httpContext, mrContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://gitlab.com/gitlab-org/gitlab/-/blob/master/README.md"
	blobContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     &[]string{"README.md"}[0],
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://gitlab.com/gitlab-org/gitlab/-/tree/test-branch?ref_type=HEAD"
	branchContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
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

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://gitlab.com/gitlab-org/gitlab/-/commits/master/?ref_type=HEADS"
	commitsContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://gitlab.com/gitlab-org/gitlab/-/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
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

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitlab.com",
		Url:    "https://gitlab.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitlab.com",
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/tree/test-branch", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitlab.com",
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal(url, "https://gitlab.com/daytonaio/daytona/-/tree/test-branch/README.md")

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/blob/main/README.md", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitlab.com",
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/commit/COMMIT_SHA", url)
}

func TestGitLabGitProvider(t *testing.T) {
	suite.Run(t, NewGitLabGitProviderTestSuite())
}
