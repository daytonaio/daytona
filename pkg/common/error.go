// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
)

var (
	ErrCtrlCAbort = errors.New("ctrl-c exit")
)

func IsCtrlCAbort(err error) bool {
	return err.Error() == ErrCtrlCAbort.Error()
}
