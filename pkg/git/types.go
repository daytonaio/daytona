// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import "time"

type GitCommitInfo struct {
	Hash      string    `json:"hash" validate:"required"`
	Author    string    `json:"author" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Message   string    `json:"message" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
} // @name GitCommitInfo
