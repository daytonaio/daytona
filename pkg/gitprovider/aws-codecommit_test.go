// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type AwsCodeCommitGitProviderTestSuite struct {
	gitProvider *AwsCodeCommitGitProvider
	suite.Suite
}

func NewAwsCodeCommitGitProviderTestSuite() *AwsCodeCommitGitProviderTestSuite {
	return &AwsCodeCommitGitProviderTestSuite{
		gitProvider: NewAwsCodeCommitGitProvider("https://ap-south-1.console.aws.amazon.com"),
	}
}

func TestAwsCodeCommitGitProvider(t *testing.T) {
	suite.Run(t, NewAwsCodeCommitGitProviderTestSuite())
}

func (g *AwsCodeCommitGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)

	repoUrl = "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/browse?region=ap-south-1"
	canHandle, _ = g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/pull-requests/1/details"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   nil,
		Sha:      nil,
		Source:   "ap-south-1.console.aws.amazon.com",
		Path:     nil,
		PrNumber: util.Pointer(uint32(1)),
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Files() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/browse/refs/heads/main/--/weew.txt"
	branchpath := "refs/heads/main/--/weew.txt"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		Source:   "ap-south-1.console.aws.amazon.com",
		Path:     &branchpath,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/browse/refs/heads/test2"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   util.Pointer("test2"),
		Source:   "ap-south-1.console.aws.amazon.com",
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/commits"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   util.Pointer("main"),
		Source:   "ap-south-1.console.aws.amazon.com",
		Sha:      nil,
		Path:     nil,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)

}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/commit/COMMIT_SHA"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Source:   "ap-south-1.console.aws.amazon.com",
		Branch:   util.Pointer("COMMIT_SHA"),
		Sha:      util.Pointer("COMMIT_SHA"),
		Path:     nil,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("demorepo"),
		Name:   util.Pointer("demorepo"),
		Owner:  util.Pointer("demorepo"),
		Source: util.Pointer("ap-south-1.console.aws.amazon.com"),
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("demorepo"),
		Name:   util.Pointer("demorepo"),
		Owner:  util.Pointer("demorepo"),
		Source: util.Pointer("ap-south-1.console.aws.amazon.com"),
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/test-branch", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("demorepo"),
		Name:   util.Pointer("demorepo"),
		Owner:  util.Pointer("demorepo"),
		Source: util.Pointer("ap-south-1.console.aws.amazon.com"),
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/test-branch/--/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/main/--/README.md", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("demorepo"),
		Name:   util.Pointer("demorepo"),
		Owner:  util.Pointer("demorepo"),
		Source: util.Pointer("ap-south-1.console.aws.amazon.com"),
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Branch: util.Pointer("COMMIT_SHA"),
		Sha:    util.Pointer("COMMIT_SHA"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/commit/COMMIT_SHA", url)
}
