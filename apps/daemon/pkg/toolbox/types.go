// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

type InitializeRequest struct {
	Token string `json:"token" binding:"required"`
} // @name InitializeRequest

type WorkDirResponse struct {
	Dir string `json:"dir"`
} // @name WorkDirResponse

type UserHomeDirResponse struct {
	Dir string `json:"dir"`
} // @name UserHomeDirResponse
