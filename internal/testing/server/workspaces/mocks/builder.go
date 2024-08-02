//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/stretchr/testify/mock"
)

var MockBuildResults = &build.BuildResult{
	User:              "test",
	ImageName:         "test",
	ProjectVolumePath: "test",
}

type MockBuilderPlugin struct {
	mock.Mock
}

type MockBuilderFactory struct {
	mock.Mock
}

func (f *MockBuilderFactory) Create(p project.Project, gpc *gitprovider.GitProviderConfig) (build.IBuilder, error) {
	return &mockBuilder{}, nil
}

func (f *MockBuilderFactory) CheckExistingBuild(p project.Project) (*build.BuildResult, error) {
	return MockBuildResults, nil
}

type mockBuilder struct {
	mock.Mock
}

func (b *mockBuilder) Build() (*build.BuildResult, error) {
	return MockBuildResults, nil
}

func (b *mockBuilder) CleanUp() error {
	return nil
}

func (b *mockBuilder) Publish() error {
	return nil
}

func (p *mockBuilder) SaveBuildResults(r build.BuildResult) error {
	args := p.Called(r)
	return args.Error(0)
}
