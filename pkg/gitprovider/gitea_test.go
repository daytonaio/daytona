// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type GiteaGitProviderTestSuite struct {
	gitProvider *GiteaGitProvider
	suite.Suite
}

func NewGiteaGitProviderTestSuite() *GiteaGitProviderTestSuite {
	return &GiteaGitProviderTestSuite{
		gitProvider: NewGiteaGitProvider("", "https://gitea.com"),
	}
}

func (g *GiteaGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://gitea.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *GiteaGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *GiteaGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://gitea.com/gitea/go-sdk/pulls/1"
	prContext := &StaticGitContext{
		Id:       "go-sdk",
		Name:     "go-sdk",
		Owner:    "gitea",
		Url:      "https://gitea.com/gitea/go-sdk.git",
		Source:   "gitea.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: util.Pointer(uint32(1)),
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(httpContext, prContext)
}

func (g *GiteaGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://gitea.com/gitea/go-sdk/src/branch/main/README.md"
	blobContext := &StaticGitContext{
		Id:       "go-sdk",
		Name:     "go-sdk",
		Owner:    "gitea",
		Url:      "https://gitea.com/gitea/go-sdk.git",
		Source:   "gitea.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     util.Pointer("README.md"),
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *GiteaGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://gitea.com/gitea/go-sdk/src/branch/test-branch"
	branchContext := &StaticGitContext{
		Id:       "go-sdk",
		Name:     "go-sdk",
		Owner:    "gitea",
		Url:      "https://gitea.com/gitea/go-sdk.git",
		Source:   "gitea.com",
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

func (g *GiteaGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://gitea.com/gitea/go-sdk/commits/branch/main"
	commitsContext := &StaticGitContext{
		Id:       "go-sdk",
		Name:     "go-sdk",
		Owner:    "gitea",
		Url:      "https://gitea.com/gitea/go-sdk.git",
		Source:   "gitea.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (g *GiteaGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://gitea.com/gitea/go-sdk/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "go-sdk",
		Name:     "go-sdk",
		Owner:    "gitea",
		Url:      "https://gitea.com/gitea/go-sdk.git",
		Source:   "gitea.com",
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

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitea.com"),
		Url:    "https://gitea.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitea.com"),
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/test-branch", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitea.com"),
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/main/README.md", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitea.com"),
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Sha:    util.Pointer("COMMIT_SHA"),
		Branch: util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/commit/COMMIT_SHA", url)
}

func TestGiteaGitProvider(t *testing.T) {
	suite.Run(t, NewGiteaGitProviderTestSuite())
}
