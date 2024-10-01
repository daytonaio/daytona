// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type AzureDevOpsGitProviderTestSuite struct {
	gitProvider *AzureDevOpsGitProvider
	suite.Suite
}

func NewAzureDevOpsGitProviderTestSuite() *AzureDevOpsGitProviderTestSuite {
	return &AzureDevOpsGitProviderTestSuite{
		gitProvider: NewAzureDevOpsGitProvider("", "https://dev.azure.com/dotslashtarun"),
	}
}

func TestAzureDevopsGitProvider(t *testing.T) {
	suite.Run(t, NewAzureDevOpsGitProviderTestSuite())
}

func (g *AzureDevOpsGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *AzureDevOpsGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1/pullrequest/4"
	prContext := &StaticGitContext{
		PrNumber: util.Pointer(uint32(4)),
		Source:   "dev.azure.com",
		Owner:    "dotslashtarun",
		Name:     "dot-1",
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Sha:      nil,
		Path:     nil,
		Branch:   nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(httpContext, prContext)
}

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1?version=GBmain"
	branchContext := &StaticGitContext{
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Name:     "dot-1",
		Owner:    "dotslashtarun",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Source:   "dev.azure.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(httpContext, branchContext)
}

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Name:     "dot-1",
		Owner:    "dotslashtarun",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Source:   "dev.azure.com",
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

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1/commits?itemVersion=GBtestbranch"
	commitContext := &StaticGitContext{
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Name:     "dot-1",
		Owner:    "dotslashtarun",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Source:   "dev.azure.com",
		Branch:   util.Pointer("testbranch"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1?path=/README.md&_a=history"
	blobContext := &StaticGitContext{
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Name:     "dot-1",
		Owner:    "dotslashtarun",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Source:   "dev.azure.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Path:     util.Pointer("README.md"),
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("dev.azure.com"),
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/daytona", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("dev.azure.com"),
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBtest-branch", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("dev.azure.com"),
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBtest-branch&path=README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBmain&path=README.md", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("dev.azure.com"),
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Branch: util.Pointer("COMMIT_SHA"),
		Sha:    util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GCCOMMIT_SHA&path=README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GCCOMMIT_SHA", url)
}
