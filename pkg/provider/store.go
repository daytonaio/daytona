// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

type Store interface {
}

type TargetStore interface {
	List() ([]*ProviderTarget, error)
	Find(providerName, targetName string) (*ProviderTarget, error)
	Save(target *ProviderTarget) error
	Delete(target *ProviderTarget) error
}
