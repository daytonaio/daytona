// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/target"
)

type TargetStore struct {
	db *gorm.DB
}

func NewTargetStore(db *gorm.DB) (*TargetStore, error) {
	err := db.AutoMigrate(&TargetDTO{})
	if err != nil {
		return nil, err
	}

	return &TargetStore{db: db}, nil
}

func (store *TargetStore) List() ([]*target.Target, error) {
	targetDTOs := []TargetDTO{}
	tx := store.db.Find(&targetDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	targets := []*target.Target{}
	for _, targetDTO := range targetDTOs {
		targets = append(targets, ToTarget(targetDTO))
	}

	return targets, nil
}

func (w *TargetStore) Find(idOrName string) (*target.Target, error) {
	targetDTO := TargetDTO{}
	tx := w.db.Where("id = ? OR name = ?", idOrName, idOrName).First(&targetDTO)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, target.ErrTargetNotFound
		}
		return nil, tx.Error
	}

	return ToTarget(targetDTO), nil
}

func (w *TargetStore) Save(target *target.Target) error {
	tx := w.db.Save(ToTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (w *TargetStore) Delete(t *target.Target) error {
	tx := w.db.Delete(ToTargetDTO(t))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return target.ErrTargetNotFound
	}

	return nil
}
