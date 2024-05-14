// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import "errors"

var (
	ErrProfileDataNotFound = errors.New("profile data not found")
)

func IsProfileDataNotFound(err error) bool {
	return err.Error() == ErrProfileDataNotFound.Error()
}
