// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"regexp"
)

func ExtractErrorPart(errorMsg string) string {
	r := regexp.MustCompile(`(unable to find user [^:]+)`)

	matches := r.FindStringSubmatch(errorMsg)

	if len(matches) < 2 {
		return errorMsg
	}

	return fmt.Sprintf("bad request: %s", matches[1])
}
