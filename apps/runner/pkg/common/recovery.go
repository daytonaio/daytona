// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/json"
	"errors"
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
// and returns the appropriate recovery type, or UnknownRecoveryType if not recoverable
func DeduceRecoveryType(errorReason string) models.RecoveryType {
	if errorReason == "" {
		return models.UnknownRecoveryType
	}

	errorReasonLower := strings.ToLower(errorReason)

	for recoveryType, patterns := range recoverableErrorPatterns {
		for _, pattern := range patterns {
			if strings.Contains(errorReasonLower, pattern) {
				return recoveryType
			}
		}
	}

	return models.UnknownRecoveryType
}

// IsRecoverable checks if an error reason is recoverable (any type)
func IsRecoverable(errorReason string) bool {
	return DeduceRecoveryType(errorReason) != models.UnknownRecoveryType
}

// Appended only on v0 HTTP; v2 omits it to preserve existing behavior.
const recoverableSuffix = " - you may attempt a recovery action"

// formatRecoverable is the v0 HTTP variant (with suffix).
func formatRecoverable(err error) (string, bool) {
	return marshalRecoverable(err, true)
}

// FormatRecoverableError is the v2 job-executor variant (no suffix).
func FormatRecoverableError(err error) error {
	if encoded, ok := marshalRecoverable(err, false); ok {
		return errors.New(encoded)
	}
	return err
}

func marshalRecoverable(err error, withSuffix bool) (string, bool) {
	msg := err.Error()
	if !IsRecoverable(msg) {
		return "", false
	}
	reason := msg
	if withSuffix {
		reason += recoverableSuffix
	}
	b, marshalErr := json.Marshal(map[string]any{
		"errorReason": reason,
		"recoverable": true,
	})
	if marshalErr != nil {
		return "", false
	}
	return string(b), true
}
