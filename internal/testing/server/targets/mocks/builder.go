//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
	"github.com/daytonaio/daytona/pkg/target/workspace/containerconfig"
	"github.com/stretchr/testify/mock"
)

var MockBuild = &build.Build{
	Id:    "1",
	State: build.BuildStatePendingRun,
	Image: util.Pointer("image"),
	User:  util.Pointer("user"),
	ContainerConfig: containerconfig.ContainerConfig{
		Image: "test",
		User:  "test",
	},
	BuildConfig: &buildconfig.BuildConfig{
		Devcontainer: MockWorkspaceConfig.BuildConfig.Devcontainer,
	},
	Repository: &gitprovider.GitRepository{
		Url: MockWorkspaceConfig.RepositoryUrl,
	},
	EnvVars: map[string]string{},
}

type MockBuilderFactory struct {
	mock.Mock
}

func (f *MockBuilderFactory) Create(build build.Build, workspaceDir string) (build.IBuilder, error) {
	args := f.Called(build, workspaceDir)
	return args.Get(0).(*MockBuilder), args.Error(1)
}

func (f *MockBuilderFactory) CheckExistingBuild(b build.Build) (*build.Build, error) {
	args := f.Called(b)
	return args.Get(0).(*build.Build), args.Error(1)
}

type MockBuilder struct {
	mock.Mock
}

func (b *MockBuilder) Build(build build.Build) (string, string, error) {
	args := b.Called(build)
	return args.String(0), args.String(1), args.Error(2)
}

func (b *MockBuilder) CleanUp() error {
	args := b.Called()
	return args.Error(0)
}

func (b *MockBuilder) Publish(build build.Build) error {
	args := b.Called(build)
	return args.Error(0)
}

func (b *MockBuilder) SaveBuild(r build.Build) error {
	args := b.Called(r)
	return args.Error(0)
}

func (b *MockBuilder) GetImageName(build build.Build) (string, error) {
	args := b.Called(build)
	return args.String(0), args.Error(1)
}
