// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"errors"
)

var (
	ErrTargetAlreadyExists = errors.New("target already exists")
	ErrInvalidTargetName   = errors.New("name is not a valid alphanumeric string")
	ErrTargetNotFound      = errors.New("target not found")
)

func IsTargetNotFound(err error) bool {
	return err.Error() == ErrTargetNotFound.Error()
}
