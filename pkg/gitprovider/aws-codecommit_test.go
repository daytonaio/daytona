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
	branch := "main"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   &branch,
		Sha:      nil,
		Source:   "ap-south-1.console.aws.amazon.com",
		Path:     nil,
		PrNumber: &[]uint32{1}[0],
	}

	require := g.Require()
	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Files() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/browse/refs/heads/main/--/weew.txt"
	branch := "main"
	branchpath := "refs/heads/main/--/weew.txt"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   &branch,
		Sha:      nil,
		Source:   "ap-south-1.console.aws.amazon.com",
		Path:     &branchpath,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/browse/refs/heads/test2"
	branch := "test2"
	branchpath := "refs/heads/test2"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   &branch,
		Sha:      nil,
		Source:   "ap-south-1.console.aws.amazon.com",
		Path:     &branchpath,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/commits"
	branch := "main"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   &branch,
		Source:   "ap-south-1.console.aws.amazon.com",
		Sha:      nil,
		Path:     nil,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)

}

func (g *AwsCodeCommitGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	prUrl := "https://ap-south-1.console.aws.amazon.com/codesuite/codecommit/repositories/Test/commit/c98e121383bc4e85d71891bdc54953187f2dd878"
	sha := "c98e121383bc4e85d71891bdc54953187f2dd878"
	prContext := &StaticGitContext{
		Id:       "Test",
		Name:     "Test",
		Owner:    "Test",
		Url:      "https://git-codecommit.ap-south-1.amazonaws.com/v1/repos/Test",
		Branch:   &sha,
		Source:   "ap-south-1.console.aws.amazon.com",
		Sha:      &sha,
		Path:     nil,
		PrNumber: nil,
	}

	require := g.Require()
	httpContext, err := g.gitProvider.parseStaticGitContext(prUrl)
	require.Nil(err)
	require.Equal(prContext, httpContext)
}
