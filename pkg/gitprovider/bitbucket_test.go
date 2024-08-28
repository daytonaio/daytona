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

	httpContext, err := b.gitProvider.ParseStaticGitContext(prUrl)

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

	httpContext, err := b.gitProvider.ParseStaticGitContext(blobUrl)

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

	httpContext, err := b.gitProvider.ParseStaticGitContext(branchUrl)

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

	httpContext, err := b.gitProvider.ParseStaticGitContext(commitsUrl)

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

	httpContext, err := b.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *BitbucketGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "bitbucket.org",
		Url:    "https://bitbucket.org/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona", url)
}

func (g *BitbucketGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "bitbucket.org",
		Url:    "https://bitbucket.org/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona/branch/test-branch", url)
}

func (g *BitbucketGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "bitbucket.org",
		Url:    "https://bitbucket.org/daytonaio/daytona.git",
		Branch: &[]string{"test-branch"}[0],
		Path:   &[]string{"README.md"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona/src/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona/src/main/README.md", url)
}

func (g *BitbucketGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GitRepository{
		Id:     "daytona",
		Name:   "daytona",
		Owner:  "daytonaio",
		Source: "bitbucket.org",
		Url:    "https://bitbucket.org/daytonaio/daytona.git",
		Path:   &[]string{"README.md"}[0],
		Sha:    "COMMIT_SHA",
		Branch: &[]string{"COMMIT_SHA"}[0],
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona/src/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromRepository(repo)

	require.Equal("https://bitbucket.org/daytonaio/daytona/commit/COMMIT_SHA", url)
}

func TestBitbucketGitProvider(t *testing.T) {
	suite.Run(t, NewBitbucketGitProviderTestSuite())
}
