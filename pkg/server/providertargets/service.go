// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providertargets

import "github.com/daytonaio/daytona/pkg/provider"

type IProviderTargetService interface {
	Delete(target *provider.ProviderTarget) error
	Find(targetName string) (*provider.ProviderTarget, error)
	List() ([]*provider.ProviderTarget, error)
	Map() (map[string]*provider.ProviderTarget, error)
	Save(target *provider.ProviderTarget) error
}

type ProviderTargetServiceConfig struct {
	TargetStore provider.TargetStore
}

type ProviderTargetService struct {
	targetStore provider.TargetStore
}

func NewProviderTargetService(config ProviderTargetServiceConfig) IProviderTargetService {
	return &ProviderTargetService{
		targetStore: config.TargetStore,
	}
}

func (s *ProviderTargetService) List() ([]*provider.ProviderTarget, error) {
	return s.targetStore.List()
}

func (s *ProviderTargetService) Map() (map[string]*provider.ProviderTarget, error) {
	list, err := s.targetStore.List()
	if err != nil {
		return nil, err
	}

	targets := make(map[string]*provider.ProviderTarget)
	for _, target := range list {
		targets[target.Name] = target
	}

	return targets, nil
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
