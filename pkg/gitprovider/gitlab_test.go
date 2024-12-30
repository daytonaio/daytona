// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
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

func (g *GitLabGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://gitlab.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *GitLabGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
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
		PrNumber: util.Pointer(uint32(1)),
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
		Branch:   util.Pointer("master"),
		Sha:      nil,
		PrNumber: nil,
		Path:     util.Pointer("README.md"),
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
		Branch:   util.Pointer("test-branch"),
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
		Branch:   util.Pointer("master"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)

	shaUrl := "https://gitlab.com/gitlab-org/gitlab/-/commits/c87cfbb77c2cf36356d010d1c0b21817c42f70ef"
	shaContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   util.Pointer("c87cfbb77c2cf36356d010d1c0b21817c42f70ef"),
		Sha:      util.Pointer("c87cfbb77c2cf36356d010d1c0b21817c42f70ef"),
		PrNumber: nil,
		Path:     nil,
	}

	httpContext, err = g.gitProvider.ParseStaticGitContext(shaUrl)
	require.Nil(err)
	require.Equal(httpContext, shaContext)
}

func (g *GitLabGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://gitlab.com/gitlab-org/gitlab/-/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "gitlab-org/gitlab",
		Name:     "gitlab",
		Owner:    "gitlab-org",
		Url:      "https://gitlab.com/gitlab-org/gitlab.git",
		Source:   "gitlab.com",
		Branch:   util.Pointer("COMMIT_SHA"),
		Sha:      util.Pointer("COMMIT_SHA"),
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitlab.com"),
		Url:    "https://gitlab.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitlab.com"),
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/tree/test-branch", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitlab.com"),
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal(url, "https://gitlab.com/daytonaio/daytona/-/tree/test-branch/README.md")

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/blob/main/README.md", url)
}

func (g *GitLabGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitlab.com"),
		Url:    "https://gitlab.com/daytonaio/daytona.git",
		Branch: util.Pointer("COMMIT_SHA"),
		Sha:    util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitlab.com/daytonaio/daytona/-/commit/COMMIT_SHA", url)
}

func TestGitLabGitProvider(t *testing.T) {
	suite.Run(t, NewGitLabGitProviderTestSuite())
}
