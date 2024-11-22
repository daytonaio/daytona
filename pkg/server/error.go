// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
)

var (
	ErrLogFileNotFound = errors.New("log file not found")
)

func IsLogFileNotFound(err error) bool {
	return err.Error() == ErrLogFileNotFound.Error()
}
