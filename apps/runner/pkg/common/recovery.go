// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"strings"

	"github.com/daytonaio/runner/pkg/models"
)

// Patterns that indicate recoverable errors mapped to their recovery types
var recoverableErrorPatterns = map[models.RecoveryType][]string{
	models.RecoveryTypeStorageExpansion: {
		"no space left on device",
		"storage limit",
		"disk quota exceeded",
	},
	// Add more recovery types here as needed:
	// dto.RecoveryTypeNetworkFailure: {
	//     "network unreachable",
	//     "connection timeout",
	// },
}

// DeduceRecoveryType determines if an error reason indicates a recoverable error
// and returns the appropriate recovery type, or empty string if not recoverable
func DeduceRecoveryType(errorReason string) models.RecoveryType {
	if errorReason == "" {
		return ""
	}

	errorReasonLower := strings.ToLower(errorReason)

	for recoveryType, patterns := range recoverableErrorPatterns {
		for _, pattern := range patterns {
			if strings.Contains(errorReasonLower, pattern) {
				return recoveryType
			}
		}
	}

	return ""
}

// IsRecoverable checks if an error reason is recoverable (any type)
func IsRecoverable(errorReason string) bool {
	return DeduceRecoveryType(errorReason) != ""
}
