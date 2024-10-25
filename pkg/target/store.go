// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import "errors"

type Store interface {
	List() ([]*Target, error)
	Find(idOrName string) (*Target, error)
	Save(target *Target) error
	Delete(target *Target) error
}

var (
	ErrTargetNotFound = errors.New("target not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
