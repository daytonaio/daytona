// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

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
	prNumber := uint32(1)
	prContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: &prNumber,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(prUrl)

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
		Path:     &[]string{"README.md"}[0],
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(blobContext, httpContext)
}

func (b *BitbucketServerGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://bitbucket.example.com/rest/api/latest/projects/PROJECT_KEY/repos/REPO_NAME/browse?at=refs/heads/master"
	branchContext := &StaticGitContext{
		Id:       "PROJECT_KEY",
		Name:     "REPO_NAME",
		Owner:    "PROJECT_KEY",
		Url:      "https://bitbucket.example.com/scm/PROJECT_KEY/REPO_NAME.git",
		Source:   "bitbucket.example.com",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(branchUrl)

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
		Branch:   nil,
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(commitsUrl)

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
		Branch:   nil,
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(commitContext, httpContext)
}

func TestBitbucketServerGitProvider(t *testing.T) {
	suite.Run(t, NewBitbucketServerGitProviderTestSuite())
}
