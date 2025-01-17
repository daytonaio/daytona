// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type TransactionKey struct{}

type IStore interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	// If an error ocurrs while rolling back the transaction, the error should be wrapped and returned,
	// otherwise, the original error is returned
	RollbackTransaction(ctx context.Context, err error) error
}

func RecoverAndRollback(ctx context.Context, store IStore) {
	if r := recover(); r != nil {
		err := store.RollbackTransaction(ctx, nil)
		if err != nil {
			// TODO: Think about this
			log.Error(err)
		}
	}
}
