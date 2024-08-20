// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type BitbucketServerGitProviderTestSuite struct {
	gitProvider *BitbucketServerGitProvider
	suite.Suite
}

func NewBitbucketServerGitProviderTestSuite() *BitbucketServerGitProviderTestSuite {
	baseApiUrl := "https://bitbucket.example.com"
	return &BitbucketServerGitProviderTestSuite{
		gitProvider: NewBitbucketServerGitProvider("username", "token", &baseApiUrl),
	}
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/pull-requests/1"
	prContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: util.Pointer(uint32(1)),
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/browse/README.md"
	blobContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Path:     util.Pointer("README.md"),
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(blobContext, httpContext)
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/browse?at=refs%2Fheads%2Fmain"
	branchContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(branchContext, httpContext)
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/commits?until=COMMIT_SHA"
	commitsContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   util.Pointer("COMMIT_SHA"),
		Sha:      util.Pointer("COMMIT_SHA"),
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(commitsContext, httpContext)
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/commits/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   util.Pointer("COMMIT_SHA"),
		Sha:      util.Pointer("COMMIT_SHA"),
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

// edge case for parsing all paths containing a anchor at the end.
func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Repo_Urls_With_Anchor() {
	commitWithAnchorUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/commits/COMMIT_SHA\\#test.txt"
	commitContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   util.Pointer("COMMIT_SHA"),
		Sha:      util.Pointer("COMMIT_SHA"),
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.ParseStaticGitContext(commitWithAnchorUrl)

	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func (g *BitbucketServerGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("bitbucket.example.com"),
		Url:    "https://bitbucket.example.com/scm/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona", url)
}

func (g *BitbucketServerGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("bitbucket.example.com"),
		Url:    "https://bitbucket.example.com/scm/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona/src/test-branch", url)
}

func (g *BitbucketServerGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("bitbucket.example.com"),
		Url:    "https://bitbucket.example.com/scm/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona/src/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona/src/main/README.md", url)
}

func (g *BitbucketServerGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("bitbucket.example.com"),
		Url:    "https://bitbucket.example.com/scm/daytonaio/daytona.git",
		Branch: util.Pointer("COMMIT_SHA"),
		Sha:    util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}
	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona/src/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://bitbucket.example.com/daytonaio/daytona/src/COMMIT_SHA", url)
}

func TestBitbucketServerGitProvider(t *testing.T) {
	suite.Run(t, NewBitbucketServerGitProviderTestSuite())
}
