// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GiteaGitProviderTestSuite struct {
	gitProvider *GiteaGitProvider
	suite.Suite
}

func NewGiteaGitProviderTestSuite() *GiteaGitProviderTestSuite {
	return &GiteaGitProviderTestSuite{
		gitProvider: NewGiteaGitProvider("", ""),
	}
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
		PrNumber: &[]uint32{1}[0],
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)

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
		Branch:   &[]string{"main"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     &[]string{"README.md"}[0],
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(blobUrl)

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
		Branch:   &[]string{"test-branch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(branchUrl)

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
		Branch:   &[]string{"main"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(commitsUrl)

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
		Branch:   &[]string{"COMMIT_SHA"}[0],
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitea.com",
		Url:    "https://gitea.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitea.com",
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/test-branch", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitea.com",
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/branch/main/README.md", url)
}

func (g *GiteaGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "gitea.com",
		Url:    "https://gitea.com/daytonaio/daytona.git",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://gitea.com/daytonaio/daytona/src/commit/COMMIT_SHA", url)
}

func TestGiteaGitProvider(t *testing.T) {
	suite.Run(t, NewGiteaGitProviderTestSuite())
}
