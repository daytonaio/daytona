// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"strings"
)

var (
	ErrCtrlCAbort = errors.New("ctrl-c exit")
)

func IsCtrlCAbort(err error) bool {
	return err.Error() == ErrCtrlCAbort.Error()
}

var (
	ErrConnection = errors.New("If you are using a VPN or firewall, please read our troubleshooting guide at https://daytona.io/docs/misc/troubleshooting#connectivity-issues")
)

func IsConnectionError(err error) bool {
	return strings.Contains(err.Error(), ErrConnection.Error())
}
