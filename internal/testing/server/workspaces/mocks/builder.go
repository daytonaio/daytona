//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/builder"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/stretchr/testify/mock"
)

var MockBuildResults = &builder.BuildResult{
	User:              "test",
	ImageName:         "test",
	ProjectVolumePath: "test",
}

type MockBuilderPlugin struct {
}

func (b *MockBuilderPlugin) Build() (*builder.BuildResult, error) {
	return MockBuildResults, nil
}

func (b *MockBuilderPlugin) CleanUp() error {
	return nil
}

func (b *MockBuilderPlugin) Publish() error {
	return nil
}

type MockBuilderFactory struct {
	mock.Mock
}

func (f *MockBuilderFactory) Create(p workspace.Project, cr *containerregistry.ContainerRegistry, gpc *gitprovider.GitProviderConfig) builder.IBuilder {
	return &mockBuilder{}
}

type mockBuilder struct {
	mock.Mock
}

func (p *mockBuilder) Prepare() error {
	args := p.Called()
	return args.Error(0)
}

func (p *mockBuilder) LoadBuildResults() (*builder.BuildResult, error) {
	args := p.Called()
	return MockBuildResults, args.Error(0)
}

func (p *mockBuilder) SaveBuildResults(r builder.BuildResult) error {
	args := p.Called(r)
	return args.Error(0)
}

func (p *mockBuilder) GetBuilderPlugin() builder.BuilderPlugin {
	plugin := &MockBuilderPlugin{}
	return plugin
}
