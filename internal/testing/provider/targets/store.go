//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"github.com/daytonaio/daytona/pkg/provider"
)

type InMemoryTargetStore struct {
	targets map[string]*provider.ProviderTarget
}

func NewInMemoryTargetStore() provider.TargetStore {
	return &InMemoryTargetStore{
		targets: make(map[string]*provider.ProviderTarget),
	}
}

func (s *InMemoryTargetStore) List() ([]*provider.ProviderTarget, error) {
	targets := []*provider.ProviderTarget{}
	for _, t := range s.targets {
		targets = append(targets, t)
	}

	return targets, nil
}

func (s *InMemoryTargetStore) Find(targetName string) (*provider.ProviderTarget, error) {
	target, ok := s.targets[targetName]
	if !ok {
		return nil, provider.ErrTargetNotFound
	}

	return target, nil
}

func (s *InMemoryTargetStore) Save(target *provider.ProviderTarget) error {
	s.targets[target.Name] = target
	return nil
}

func (s *InMemoryTargetStore) Delete(target *provider.ProviderTarget) error {
	delete(s.targets, target.Name)
	return nil
}
