// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"errors"
	"time"
)

var (
	ErrShellNotFound = errors.New("no shell found")
)

type ErrorResponse struct {
	StatusCode int       `json:"statusCode" example:"400" binding:"required"`
	Message    string    `json:"message" example:"Bad request" binding:"required"`
	Code       string    `json:"code" example:"BAD_REQUEST" binding:"required"`
	Timestamp  time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z" binding:"required"`
	Path       string    `json:"path" example:"/api/resource" binding:"required"`
	Method     string    `json:"method" example:"GET" binding:"required"`
} //	@name	ErrorResponse
