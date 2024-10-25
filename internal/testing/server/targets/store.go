//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import "github.com/daytonaio/daytona/pkg/target"

type InMemoryTargetStore struct {
	targets map[string]*target.Target
}

func NewInMemoryTargetStore() target.Store {
	return &InMemoryTargetStore{
		targets: make(map[string]*target.Target),
	}
}

func (s *InMemoryTargetStore) List() ([]*target.Target, error) {
	targets := []*target.Target{}
	for _, t := range s.targets {
		targets = append(targets, t)
	}

	return targets, nil
}

func (s *InMemoryTargetStore) Find(idOrName string) (*target.Target, error) {
	t, ok := s.targets[idOrName]
	if !ok {
		for _, w := range s.targets {
			if w.Name == idOrName {
				return w, nil
			}
		}
		return nil, target.ErrTargetNotFound
	}

	return t, nil
}

func (s *InMemoryTargetStore) Save(target *target.Target) error {
	s.targets[target.Id] = target
	return nil
}

func (s *InMemoryTargetStore) Delete(target *target.Target) error {
	delete(s.targets, target.Id)
	return nil
}
