// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import "gorm.io/gorm"

func IsRecordNotFound(err error) bool {
	return err.Error() == gorm.ErrRecordNotFound.Error()
}
