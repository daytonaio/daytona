// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providertargets

import "github.com/daytonaio/daytona/pkg/provider"

type ProviderTargetServiceConfig struct {
	TargetStore provider.TargetStore
}

type ProviderTargetService struct {
	targetStore provider.TargetStore
}

func NewProviderTargetService(config ProviderTargetServiceConfig) *ProviderTargetService {
	return &ProviderTargetService{
		targetStore: config.TargetStore,
	}
}

func (s *ProviderTargetService) List() ([]*provider.ProviderTarget, error) {
	return s.targetStore.List()
}

func (s *ProviderTargetService) Find(targetName string) (*provider.ProviderTarget, error) {
	return s.targetStore.Find(targetName)
}

func (s *ProviderTargetService) Save(target *provider.ProviderTarget) error {
	return s.targetStore.Save(target)
}

func (s *ProviderTargetService) Delete(target *provider.ProviderTarget) error {
	return s.targetStore.Delete(target)
}
