// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import "errors"

type TargetFilter struct {
	Name    *string
	Default *bool
}

type TargetStore interface {
	List(filter *TargetFilter) ([]*ProviderTarget, error)
	Find(filter *TargetFilter) (*ProviderTarget, error)
	Save(target *ProviderTarget) error
	Delete(target *ProviderTarget) error
}

var (
	ErrTargetNotFound = errors.New("target not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
