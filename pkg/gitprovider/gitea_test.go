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

func TestGiteaGitProvider(t *testing.T) {
	suite.Run(t, NewGiteaGitProviderTestSuite())
}
