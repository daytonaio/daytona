// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type GiteeGitProviderTestSuite struct {
	gitProvider *GiteeGitProvider
	suite.Suite
}

func NewGiteeGitProviderTestSuite() *GiteeGitProviderTestSuite {
	return &GiteeGitProviderTestSuite{
		gitProvider: NewGiteeGitProvider(""),
	}
}

func (g *GiteeGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://gitee.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *GiteeGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *GiteeGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://gitee.com/daytonaio/daytona/pulls/1"
	prContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gitee.com/daytonaio/daytona.git",
		Source:   "gitee.com",
		Branch:   nil,
		Sha:      nil,
		PrNumber: util.Pointer(uint32(1)),
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(prUrl)

	require.Nil(err)
	require.Equal(httpContext, prContext)
}

func (g *GiteeGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://gitee.com/daytonaio/daytona/blob/fff62f9717b0e4d2f9262d159b90f24efc626021/README.en.md"
	blobContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gitee.com/daytonaio/daytona.git",
		Source:   "gitee.com",
		Branch:   util.Pointer("fff62f9717b0e4d2f9262d159b90f24efc626021"),
		Sha:      util.Pointer("fff62f9717b0e4d2f9262d159b90f24efc626021"),
		PrNumber: nil,
		Path:     util.Pointer("README.en.md"),
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *GiteeGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://gitee.com/daytonaio/daytona/tree/test"
	branchContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gitee.com/daytonaio/daytona.git",
		Source:   "gitee.com",
		Branch:   util.Pointer("test"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(branchUrl)

	require.Nil(err)
	require.Equal(httpContext, branchContext)
}

func (g *GiteeGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://gitee.com/daytonaio/daytona/commits/master"
	commitsContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gitee.com/daytonaio/daytona.git",
		Source:   "gitee.com",
		Branch:   util.Pointer("master"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (g *GiteeGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://gitee.com/daytonaio/daytona/commit/fff62f9717b0e4d2f9262d159b90f24efc626021"
	commitContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gitee.com/daytonaio/daytona.git",
		Source:   "gitee.com",
		Branch:   util.Pointer("fff62f9717b0e4d2f9262d159b90f24efc626021"),
		Sha:      util.Pointer("fff62f9717b0e4d2f9262d159b90f24efc626021"),
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitUrl)

	require.Nil(err)
	require.Equal(httpContext, commitContext)
}

func (g *GiteeGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitee.com"),
		Url:    "https://gitee.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitee.com/daytonaio/daytona", url)
}

func (g *GiteeGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gitee.com"),
		Url:    "https://gitee.com/daytonaio/daytona.git",
		Branch: util.Pointer("test"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gitee.com/daytonaio/daytona/tree/test", url)
}

func TestGiteeGitProvider(t *testing.T) {
	suite.Run(t, NewGiteeGitProviderTestSuite())
}
