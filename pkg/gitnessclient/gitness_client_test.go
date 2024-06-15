// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitnessclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GitnessClientTestSuite struct {
	suite.Suite
	client *GitnessClient
}

func (suite *GitnessClientTestSuite) SetupSuite() {
	baseUrl, _ := url.Parse("http://localhost:3000/api/v1")
	token := ""
	suite.client = NewGitnessClient(token, baseUrl)
}

func (suite *GitnessClientTestSuite) TestGetSpaceAdmin() {
	require := suite.Require()
	admin, err := suite.client.GetSpaceAdmin("test")

	require.Nil(err)
	require.NotNil(admin)
	// require.Equal("space_owner", admin.Role)
}

func (suite *GitnessClientTestSuite) TestGetSpaces() {
	require := suite.Require()

	spaces, err := suite.client.GetSpaces()

	require.Nil(err)
	require.NotEmpty(spaces)
}

func (suite *GitnessClientTestSuite) TestGetUser() {
	require := suite.Require()

	user, err := suite.client.GetUser()

	require.Nil(err)
	require.NotNil(user)
}

func (suite *GitnessClientTestSuite) TestGetRepositories() {
	require := suite.Require()

	repos, err := suite.client.GetRepositories("test")

	require.Nil(err)
	require.NotEmpty(repos)
}

func (suite *GitnessClientTestSuite) TestGetRepoBranches() {
	require := suite.Require()

	branches, err := suite.client.GetRepoBranches("test", "test")

	require.Nil(err)
	require.NotEmpty(branches)
}

func (suite *GitnessClientTestSuite) TestGetRepoPRs() {
	require := suite.Require()

	prs, err := suite.client.GetRepoPRs("test", "test")

	require.Nil(err)
	require.NotEmpty(prs)
}

func (suite *GitnessClientTestSuite) TestGetLastCommitSha() {
	require := suite.Require()
	branch := "main"
	sha, err := suite.client.GetLastCommitSha("http://localhost:3000/test/test", &branch)

	require.Nil(err)
	require.NotEmpty(sha)
}

func (suite *GitnessClientTestSuite) TestGetPr() {
	require := suite.Require()
	pr, err := suite.client.GetPr("http://localhost:3000/test/test", 1)
	require.Nil(err)
	require.NotNil(pr)
}
func (suite *GitnessClientTestSuite) TestWithDifferentUrl() {
	urls := []string{
		"http://localhost:3000/test/test",
		"http://localhost:3000/test/test/commits/testbranch",
		"http://localhost:3000/test/test/commits/main",
		"http://localhost:3000/test/test/branches",
		"http://localhost:3000/test/test/pulls?state=open",
		"http://localhost:3000/git/test/test.git",
	}

	for _, url := range urls {
		suite.Run(url, func() {

			require := suite.Require()

			branch := "main"
			testBranch := "testbranch"

			sha, err := suite.client.GetLastCommitSha(url, &branch)
			require.Nil(err, "failed to get last commit SHA for URL: %s", url)
			require.NotEmpty(sha, "commit SHA should not be empty for URL: %s", url)

			shaTestBranch, err := suite.client.GetLastCommitSha(url, &testBranch)
			require.Nil(err, "failed to get last commit SHA for URL: %s", url)
			require.NotEmpty(shaTestBranch, "commit SHA should not be empty for URL: %s", url)

			pr, err := suite.client.GetPr(url, 1)
			require.Nil(err, "failed to get pull request for URL: %s", url)
			require.NotNil(pr, "pull request should not be nil for URL: %s", url)

		})
	}
}
func TestGitnessClient(t *testing.T) {
	suite.Run(t, new(GitnessClientTestSuite))
}
