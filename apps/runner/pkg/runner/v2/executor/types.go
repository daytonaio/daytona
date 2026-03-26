/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

type StartSandboxPayload struct {
	AuthToken *string           `json:"authToken,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}
