// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import "errors"

type Store interface {
}

type TargetStore interface {
	List() ([]*ProviderTarget, error)
	Find(targetName string) (*ProviderTarget, error)
	Save(target *ProviderTarget) error
	Delete(target *ProviderTarget) error
}

var (
	ErrTargetNotFound = errors.New("provider not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
