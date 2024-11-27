// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type ProfileDataStore struct {
	db *gorm.DB
}

func NewProfileDataStore(db *gorm.DB) (*ProfileDataStore, error) {
	err := db.AutoMigrate(&models.ProfileData{})
	if err != nil {
		return nil, err
	}

	return &ProfileDataStore{db: db}, nil
}

func (p *ProfileDataStore) Get(id string) (*models.ProfileData, error) {
	profileData := &models.ProfileData{}
	tx := p.db.Where("id = ?", id).First(profileData)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, stores.ErrProfileDataNotFound
		}
		return nil, tx.Error
	}

	return profileData, nil
}

func (p *ProfileDataStore) Save(profileData *models.ProfileData) error {
	tx := p.db.Save(profileData)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (p *ProfileDataStore) Delete(id string) error {
	tx := p.db.Where("id = ?", id).Delete(&models.ProfileData{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrProfileDataNotFound
	}

	return nil
}