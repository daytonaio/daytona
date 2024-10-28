//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/stretchr/testify/mock"
)

type mockApiKeyService struct {
	mock.Mock
}

func NewMockApiKeyService() *mockApiKeyService {
	return &mockApiKeyService{}
}

func (s *mockApiKeyService) Generate(keyType apikey.ApiKeyType, name string) (string, error) {
	args := s.Called(keyType, name)
	return args.String(0), args.Error(1)
}

func (s *mockApiKeyService) IsWorkspaceApiKey(apiKey string) bool {
	args := s.Called(apiKey)
	return args.Bool(0)
}

func (s *mockApiKeyService) IsValidApiKey(apiKey string) bool {
	args := s.Called(apiKey)
	return args.Bool(0)
}

func (s *mockApiKeyService) IsTargetApiKey(apiKey string) bool {
	args := s.Called(apiKey)
	return args.Bool(0)
}

func (s *mockApiKeyService) ListClientKeys() ([]*apikey.ApiKey, error) {
	args := s.Called()
	return args.Get(0).([]*apikey.ApiKey), args.Error(1)
}

func (s *mockApiKeyService) Revoke(name string) error {
	args := s.Called(name)
	return args.Error(0)
}
