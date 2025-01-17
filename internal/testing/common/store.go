//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import "context"

type InMemoryStore struct{}

func (s *InMemoryStore) BeginTransaction(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *InMemoryStore) CommitTransaction(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) RollbackTransaction(ctx context.Context, err error) error {
	return err
}
