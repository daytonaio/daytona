// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import "errors"

type TargetFilter struct {
	IdOrName *string
	Default  *bool
}

type Store interface {
	List(filter *TargetFilter) ([]*TargetViewDTO, error)
	Find(filter *TargetFilter) (*TargetViewDTO, error)
	Save(target *Target) error
	Delete(target *Target) error
}

type TargetViewDTO struct {
	Target
	WorkspaceCount int `json:"workspaceCount" validate:"required"`
} // @name TargetViewDTO

var (
	ErrTargetNotFound = errors.New("target not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
