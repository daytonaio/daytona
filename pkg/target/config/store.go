// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "errors"

type TargetConfigFilter struct {
	Name *string
}

type TargetConfigStore interface {
	List(filter *TargetConfigFilter) ([]*TargetConfig, error)
	Find(filter *TargetConfigFilter) (*TargetConfig, error)
	Save(targetConfig *TargetConfig) error
	Delete(targetConfig *TargetConfig) error
}

var (
	ErrTargetConfigNotFound = errors.New("target config not found")
)

func IsTargetConfigNotFound(err error) bool {
	return err.Error() == ErrTargetConfigNotFound.Error()
}
