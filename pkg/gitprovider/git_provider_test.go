// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AbstractGitProviderTestSuite struct {
	*AbstractGitProvider
	suite.Suite
}

func NewAbstractGitProviderTestSuite() *AbstractGitProviderTestSuite {
	return &AbstractGitProviderTestSuite{}
}

func (a *AbstractGitProviderTestSuite) TestParseGitContext_HTTP() {
	httpSimple := "https://github.com/daytonaio/daytona.git"
	simpleContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Source:   "github.com",
		Path:     nil,
	}

	require := a.Require()

	httpContext, err := a.AbstractGitProvider.ParseStaticGitContext(httpSimple)

	require.Nil(err)
	require.Equal(httpContext, simpleContext)
}

func (a *AbstractGitProviderTestSuite) TestParseGitContext_SSH() {
	sshSimple := "git@github.com:daytonaio/daytona.git"
	simpleContext := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Source:   "github.com",
		Path:     nil,
	}

	require := a.Require()

	sshContext, err := a.AbstractGitProvider.ParseStaticGitContext(sshSimple)

	require.Nil(err)
	require.Equal(sshContext, simpleContext)
}

func (a *AbstractGitProviderTestSuite) TestParseGitContext_HTTPWithPath() {
	httpWithPath := "https://github.com/daytonaio/daytona/blob/main/README.md"
	contextWithPath := &StaticGitContext{
		Id:       "daytona",
		Name:     "daytona",
		Owner:    "daytonaio",
		Url:      "https://github.com/daytonaio/daytona.git",
		Branch:   nil,
		Sha:      nil,
		PrNumber: nil,
		Source:   "github.com",
		Path:     &[]string{"blob/main/README.md"}[0],
	}

	require := a.Require()

	httpContext, err := a.AbstractGitProvider.ParseStaticGitContext(httpWithPath)

	require.Nil(err)
	require.Equal(httpContext, contextWithPath)
}

func TestAbstractGitProvider(t *testing.T) {
	suite.Run(t, NewAbstractGitProviderTestSuite())
}
