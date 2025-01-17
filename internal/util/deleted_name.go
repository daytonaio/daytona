// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/constants"
)

func AddDeletedToName(name string) string {
	return fmt.Sprintf("%s%s%s", constants.DELETED_CIRCUMFIX, name, constants.DELETED_CIRCUMFIX)
}
