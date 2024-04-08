// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/provider"
)

type ProviderTargetStore struct {
	db *gorm.DB
}

func NewProviderTargetStore(db *gorm.DB) (*ProviderTargetStore, error) {
	err := db.AutoMigrate(&ProviderTargetDTO{})
	if err != nil {
		return nil, err
	}

	return &ProviderTargetStore{db: db}, nil
}

func (s *ProviderTargetStore) List() ([]*provider.ProviderTarget, error) {
	providerTargetsDTOs := []ProviderTargetDTO{}
	tx := s.db.Find(&providerTargetsDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	providerTargets := []*provider.ProviderTarget{}
	for _, providerTargetDTO := range providerTargetsDTOs {
		providerTargets = append(providerTargets, ToProviderTarget(providerTargetDTO))
	}

	return providerTargets, nil
}

func (s *ProviderTargetStore) Find(targetName string) (*provider.ProviderTarget, error) {
	providerTargetDTO := ProviderTargetDTO{}
	tx := s.db.Where("name = ?", targetName).First(&providerTargetDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return ToProviderTarget(providerTargetDTO), nil
}

func (s *ProviderTargetStore) Save(target *provider.ProviderTarget) error {
	tx := s.db.Save(ToProviderTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *ProviderTargetStore) Delete(target *provider.ProviderTarget) error {
	tx := s.db.Delete(ToProviderTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
