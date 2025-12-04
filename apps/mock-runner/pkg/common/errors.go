// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// HandlePossibleDockerError converts errors to HTTP error responses
// For mock runner, we don't have real Docker errors, so just return generic errors
func HandlePossibleDockerError(ctx *gin.Context, err error) common_errors.ErrorResponse {
	return common_errors.ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Code:       "INTERNAL_SERVER_ERROR",
		Timestamp:  time.Now(),
		Path:       ctx.Request.URL.Path,
		Method:     ctx.Request.Method,
	}
}
