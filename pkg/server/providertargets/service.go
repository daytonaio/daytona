// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providertargets

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/provider"
)

type IProviderTargetService interface {
	Delete(target *provider.ProviderTarget) error
	Find(filter *provider.TargetFilter) (*provider.ProviderTarget, error)
	List(filter *provider.TargetFilter) ([]*provider.ProviderTarget, error)
	Map() (map[string]*provider.ProviderTarget, error)
	Save(target *provider.ProviderTarget) error
	SetDefault(target *provider.ProviderTarget) error
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

func (s *ProviderTargetService) List(filter *provider.TargetFilter) ([]*provider.ProviderTarget, error) {
	return s.targetStore.List(filter)
}

func (s *ProviderTargetService) Map() (map[string]*provider.ProviderTarget, error) {
	list, err := s.targetStore.List(nil)
	if err != nil {
		return nil, err
	}

	targets := make(map[string]*provider.ProviderTarget)
	for _, target := range list {
		targets[target.Name] = target
	}

	return targets, nil
}

func (s *ProviderTargetService) Find(filter *provider.TargetFilter) (*provider.ProviderTarget, error) {
	return s.targetStore.Find(filter)
}

func (s *ProviderTargetService) Save(target *provider.ProviderTarget) error {
	err := s.targetStore.Save(target)
	if err != nil {
		return err
	}

	return s.SetDefault(target)
}

func (s *ProviderTargetService) Delete(target *provider.ProviderTarget) error {
	return s.targetStore.Delete(target)
}

func (s *ProviderTargetService) SetDefault(target *provider.ProviderTarget) error {
	currentTarget, err := s.Find(&provider.TargetFilter{
		Name: &target.Name,
	})
	if err != nil {
		return err
	}

	defaultTarget, err := s.Find(&provider.TargetFilter{
		Default: util.Pointer(true),
	})
	if err != nil && err != provider.ErrTargetNotFound {
		return err
	}

	if defaultTarget != nil {
		defaultTarget.IsDefault = false
		err := s.targetStore.Save(defaultTarget)
		if err != nil {
			return err
		}
	}

	currentTarget.IsDefault = true
	return s.targetStore.Save(currentTarget)
}
