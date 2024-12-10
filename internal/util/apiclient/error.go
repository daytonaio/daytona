// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"fmt"
	"strings"
)

func ErrHealthCheckFailed(healthUrl string) error {
	return fmt.Errorf("failed to check server health at: %s. Make sure all Daytona services are running on the appropriate ports", healthUrl)
}

func IsHealthCheckFailed(err error) bool {
	return strings.HasPrefix(err.Error(), "failed to check server health at:")
}
