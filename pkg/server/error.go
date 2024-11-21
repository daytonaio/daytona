// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
)

var (
	ErrLogFileDoesntExists = errors.New("log file does not exist")
)

func IsLogFileDoesntExists(err error) bool {
	return err.Error() == ErrLogFileDoesntExists.Error()
}
