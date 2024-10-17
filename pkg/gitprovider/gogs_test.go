// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/stretchr/testify/suite"
)

type GogsGitProviderTestSuite struct {
	gitProvider *GogsGitProvider
	suite.Suite
}

func NewGogsGitProviderTestSuite() *GogsGitProviderTestSuite {
	return &GogsGitProviderTestSuite{
		gitProvider: NewGogsGitProvider("", "https://gogs-host.com"),
	}
}

func (g *GogsGitProviderTestSuite) TestCanHandle() {
	repoUrl := "https://gogs-host.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.True(canHandle)
}

func (g *GogsGitProviderTestSuite) TestCanHandle_False() {
	repoUrl := "https://github.com/daytonaio/daytona"
	require := g.Require()
	canHandle, _ := g.gitProvider.CanHandle(repoUrl)
	require.False(canHandle)
}

func (g *GogsGitProviderTestSuite) TestParseStaticGitContext_PR() {
	prUrl := "https://gogs-host.com/daytonaio/daytona/pulls/1"
	prContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gogs-host.com/daytonaio/daytona.git",
		Source:   "gogs-host.com",
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

func (g *GogsGitProviderTestSuite) TestParseStaticGitContext_Blob() {
	blobUrl := "https://gogs-host.com/daytonaio/daytona/src/main/README.md"
	blobContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gogs-host.com/daytonaio/daytona.git",
		Source:   "gogs-host.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     util.Pointer("README.md"),
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(blobUrl)

	require.Nil(err)
	require.Equal(httpContext, blobContext)
}

func (g *GogsGitProviderTestSuite) TestParseStaticGitContext_Branch() {
	branchUrl := "https://gogs-host.com/daytonaio/daytona/src/main"
	branchContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gogs-host.com/daytonaio/daytona.git",
		Source:   "gogs-host.com",
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

func (g *GogsGitProviderTestSuite) TestParseStaticGitContext_Commits() {
	commitsUrl := "https://gogs-host.com/daytonaio/daytona/commits/main"
	commitsContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gogs-host.com/daytonaio/daytona.git",
		Source:   "gogs-host.com",
		Branch:   util.Pointer("main"),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}

	require := g.Require()

	httpContext, err := g.gitProvider.ParseStaticGitContext(commitsUrl)

	require.Nil(err)
	require.Equal(httpContext, commitsContext)
}

func (g *GogsGitProviderTestSuite) TestParseStaticGitContext_Commit() {
	commitUrl := "https://gogs-host.com/daytonaio/daytona/commit/COMMIT_SHA"
	commitContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://gogs-host.com/daytonaio/daytona.git",
		Source:   "gogs-host.com",
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

func (g *GogsGitProviderTestSuite) TestGetUrlFromRepo_Bare() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gogs-host.com"),
		Url:    "https://gogs-host.com/daytonaio/daytona.git",
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona", url)
}

func (g *GogsGitProviderTestSuite) TestGetUrlFromRepo_Branch() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gogs-host.com"),
		Url:    "https://gogs-host.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona/src/test-branch", url)
}

func (g *GogsGitProviderTestSuite) TestGetUrlFromRepo_Path() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gogs-host.com"),
		Url:    "https://gogs-host.com/daytonaio/daytona.git",
		Branch: util.Pointer("test-branch"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona/src/test-branch/README.md", url)

	repo.Branch = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona/src/main/README.md", url)
}

func (g *GogsGitProviderTestSuite) TestGetUrlFromRepo_Commit() {
	repo := &GetRepositoryContext{
		Id:     util.Pointer("daytona"),
		Name:   util.Pointer("daytona"),
		Owner:  util.Pointer("daytonaio"),
		Source: util.Pointer("gogs-host.com"),
		Url:    "https://gogs-host.com/daytonaio/daytona.git",
		Sha:    util.Pointer("COMMIT_SHA"),
		Branch: util.Pointer("COMMIT_SHA"),
		Path:   util.Pointer("README.md"),
	}

	require := g.Require()

	url := g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona/commit/COMMIT_SHA/README.md", url)

	repo.Path = nil

	url = g.gitProvider.GetUrlFromContext(repo)

	require.Equal("https://gogs-host.com/daytonaio/daytona/commit/COMMIT_SHA", url)
}

func TestGogsGitProvider(t *testing.T) {
	suite.Run(t, NewGogsGitProviderTestSuite())
}
