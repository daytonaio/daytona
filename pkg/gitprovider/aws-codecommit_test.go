// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

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
		PrNumber: &[]uint32{1}[0],
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
		Branch:   &[]string{"main"}[0],
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
		Branch:   &[]string{"test2"}[0],
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
		Branch:   &[]string{"main"}[0],
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
		Branch:   &[]string{"COMMIT_SHA"}[0],
		Sha:      &[]string{"COMMIT_SHA"}[0],
		Path:     nil,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "demorepo",
		Name:   "demorepo",
		Owner:  "demorepo",
		Source: "ap-south-1.console.aws.amazon.com",
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "demorepo",
		Name:   "demorepo",
		Owner:  "demorepo",
		Source: "ap-south-1.console.aws.amazon.com",
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/test-branch", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "demorepo",
		Name:   "demorepo",
		Owner:  "demorepo",
		Source: "ap-south-1.console.aws.amazon.com",
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/test-branch/--/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/browse/refs/heads/main/--/README.md", url)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "demorepo",
		Name:   "demorepo",
		Owner:  "demorepo",
		Source: "ap-south-1.console.aws.amazon.com",
		Url:    "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/demorepo",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/commit/COMMIT_SHA", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/demorepo/commit/COMMIT_SHA", url)
}
