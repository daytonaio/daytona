// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/stores"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return Store{db}
}

func (s *Store) BeginTransaction(ctx context.Context) (context.Context, error) {
	if ctx.Value(stores.TransactionKey{}) != nil {
		return ctx, nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	ctx = context.WithValue(ctx, stores.TransactionKey{}, tx)
	return ctx, nil
}

func (s *Store) CommitTransaction(ctx context.Context) error {
	tx, ok := ctx.Value(stores.TransactionKey{}).(*gorm.DB)
	if !ok {
		return nil
	}

	return tx.Commit().Error
}

func (s *Store) RollbackTransaction(ctx context.Context, err error) error {
	tx, ok := ctx.Value(stores.TransactionKey{}).(*gorm.DB)
	if !ok {
		return err
	}

	txErr := tx.Rollback().Error
	if txErr == nil {
		return err
	}

	return fmt.Errorf("%w. Transaction rollback error: %w", err, txErr)
}

func (s *Store) getTransaction(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(stores.TransactionKey{}).(*gorm.DB)
	if !ok {
		return s.db
	}

	return tx
}
