//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/stretchr/testify/mock"
)

var MockBuild = &build.Build{
	Id:      "1",
	Hash:    "test",
	Project: MockProject,
	State:   build.BuildStatePending,
	User:    "test",
	Image:   "test",
}

type MockBuilderPlugin struct {
	mock.Mock
}

type MockBuilderFactory struct {
	mock.Mock
}

func (f *MockBuilderFactory) Create(build build.Build) (build.IBuilder, error) {
	return &mockBuilder{}, nil
}

func (f *MockBuilderFactory) CheckExistingBuild(p project.Project) (*build.Build, error) {
	return MockBuild, nil
}

type mockBuilder struct {
	mock.Mock
}

func (b *mockBuilder) Build(build build.Build) (string, string, error) {
	return "image", "user", nil
}

func (b *mockBuilder) CleanUp() error {
	return nil
}

func (b *mockBuilder) Publish(build build.Build) error {
	return nil
}

func (p *mockBuilder) SaveBuild(r build.Build) error {
	args := p.Called(r)
	return args.Error(0)
}
