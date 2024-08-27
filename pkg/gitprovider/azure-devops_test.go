// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

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

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1/pullrequest/4"
	prContext := &StaticGitContext{
		PrNumber: &[]uint32{4}[0],
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

	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)

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
		Branch:   &[]string{"main"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(branchUrl)

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

func (g *AzureDevOpsGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitUrl := "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1/commits?itemVersion=GBtestbranch"
	commitContext := &StaticGitContext{
		Id:       "89cfa7f7-a58b-42df-b7f2-2d11ac676a1e",
		Name:     "dot-1",
		Owner:    "dotslashtarun",
		Url:      "https://dev.azure.com/dotslashtarun/dot-1/_git/dot-1",
		Source:   "dev.azure.com",
		Branch:   &[]string{"testbranch"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(commitUrl)

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
		Path:     &[]string{"README.md"}[0],
	}

	require := g.Require()

	httpContext, err := g.gitProvider.parseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "dev.azure.com",
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/daytona", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "dev.azure.com",
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBtest-branch", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "dev.azure.com",
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBtest-branch&path=README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GBmain&path=README.md", url)
}

func (g *AzureDevOpsGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "dev.azure.com",
		Url:    "https://dev.azure.com/daytonaio/daytona.git",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GCCOMMIT_SHA&path=README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://dev.azure.com/daytonaio/_git/daytona?version=GCCOMMIT_SHA", url)
}
