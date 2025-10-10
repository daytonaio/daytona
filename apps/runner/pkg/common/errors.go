// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/runner/internal/util"
	"github.com/docker/docker/errdefs"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func HandlePossibleDockerError(ctx *gin.Context, err error) common_errors.ErrorResponse {
	if errdefs.IsUnauthorized(err) || strings.Contains(err.Error(), "unauthorized") {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    fmt.Sprintf("unauthorized: %s", err.Error()),
			Code:       "UNAUTHORIZED",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsConflict(err) {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusConflict,
			Message:    fmt.Sprintf("conflict: %s", err.Error()),
			Code:       "CONFLICT",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsInvalidParameter(err) {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("bad request: %s", err.Error()),
			Code:       "BAD_REQUEST",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsSystem(err) {
		if strings.Contains(err.Error(), "unable to find user") {
			return common_errors.ErrorResponse{
				StatusCode: http.StatusBadRequest,
				Message:    util.ExtractErrorPart(err.Error()),
				Code:       "BAD_REQUEST",
				Timestamp:  time.Now(),
				Path:       ctx.Request.URL.Path,
				Method:     ctx.Request.Method,
			}
		}
	}

	return common_errors.ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Code:       "INTERNAL_SERVER_ERROR",
		Timestamp:  time.Now(),
		Path:       ctx.Request.URL.Path,
		Method:     ctx.Request.Method,
	}
}
