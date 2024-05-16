// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BitbucketGitProviderTestSuite struct {
	gitProvider *BitbucketGitProvider
	suite.Suite
}

func NewBitbucketGitProviderTestSuite() *BitbucketGitProviderTestSuite {
	return &BitbucketGitProviderTestSuite{
		gitProvider: NewBitbucketGitProvider("", ""),
	}
}

func (b *BitbucketGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://bitbucket.org/atlassian/bitbucket-upload-file/pull-requests/1"
	prContext := &StaticGitContext{
		Id:       "bitbucket-upload-file",
		Name:     "bitbucket-upload-file",
		Owner:    "atlassian",
		Url:      "https://bitbucket.org/atlassian/bitbucket-upload-file.git",
		Source:   "bitbucket.org",
		Branch:   nil,
		Sha:      nil,
		PrNumber: &[]uint32{1}[0],
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(httpContext, prContext)
}

func (b *BitbucketGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://bitbucket.org/atlassian/bitbucket-upload-file/src/master/README.md"
	blobContext := &StaticGitContext{
		Id:       "bitbucket-upload-file",
		Name:     "bitbucket-upload-file",
		Owner:    "atlassian",
		Url:      "https://bitbucket.org/atlassian/bitbucket-upload-file.git",
		Source:   "bitbucket.org",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     &[]string{"README.md"}[0],
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (b *BitbucketGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://bitbucket.org/atlassian/bitbucket-upload-file/src/master"
	branchContext := &StaticGitContext{
		Id:       "bitbucket-upload-file",
		Name:     "bitbucket-upload-file",
		Owner:    "atlassian",
		Url:      "https://bitbucket.org/atlassian/bitbucket-upload-file.git",
		Source:   "bitbucket.org",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(httpContext, branchContext)
}

func (b *BitbucketGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://bitbucket.org/atlassian/bitbucket-upload-file/commits/branch/master"
	commitsContext := &StaticGitContext{
		Id:       "bitbucket-upload-file",
		Name:     "bitbucket-upload-file",
		Owner:    "atlassian",
		Url:      "https://bitbucket.org/atlassian/bitbucket-upload-file.git",
		Source:   "bitbucket.org",
		Branch:   &[]string{"master"}[0],
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (b *BitbucketGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://bitbucket.org/atlassian/bitbucket-upload-file/commits/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "bitbucket-upload-file",
		Name:     "bitbucket-upload-file",
		Owner:    "atlassian",
		Url:      "https://bitbucket.org/atlassian/bitbucket-upload-file.git",
		Source:   "bitbucket.org",
		Branch:   &[]string{"COMMIT_SHA"}[0],
		Sha:      &[]string{"COMMIT_SHA"}[0],
		PrNumber: nil,
		Path:     nil,
	}

	require := b.Require()

	httpContext, err := b.gitProvider.parseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func TestBitbucketGitProvider(t *testing.T) {
	suite.Run(t, NewBitbucketGitProviderTestSuite())
}
