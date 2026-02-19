// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

type InitializeRequest struct {
	Token string `json:"token" binding:"required"`
} // @name InitializeRequest

type WorkDirResponse struct {
	Dir string `json:"dir" binding:"required"`
} // @name WorkDirResponse

type UserHomeDirResponse struct {
	Dir string `json:"dir" binding:"required"`
} // @name UserHomeDirResponse
