// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package process

type ExecuteRequest struct {
	Command string `json:"command" validate:"required"`
	// Timeout in seconds, defaults to 10 seconds
	Timeout *uint32 `json:"timeout,omitempty" validate:"optional"`
} // @name ExecuteRequest

type ExecuteResponse struct {
	Code   int    `json:"code" validate:"required"`
	Result string `json:"result" validate:"required"`
} // @name ExecuteResponse
