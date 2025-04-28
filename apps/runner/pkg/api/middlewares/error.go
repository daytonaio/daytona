// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/errdefs"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		errs := ctx.Errors
		if len(errs) > 0 {
			var errorResponse common.ErrorResponse
			err := errs.Last()

			switch e := err.Err.(type) {
			case *common.CustomError:
				errorResponse = common.ErrorResponse{
					StatusCode: e.StatusCode,
					Message:    e.Message,
					Code:       e.Code,
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *common.NotFoundError:
				errorResponse = common.ErrorResponse{
					StatusCode: http.StatusNotFound,
					Message:    err.Err.Error(),
					Code:       "NOT_FOUND",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *common.UnauthorizedError:
				errorResponse = common.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    err.Err.Error(),
					Code:       "UNAUTHORIZED",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *common.InvalidBodyRequestError:
				errorResponse = common.ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    err.Err.Error(),
					Code:       "INVALID_REQUEST_BODY",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *common.ConflictError:
				errorResponse = common.ErrorResponse{
					StatusCode: http.StatusConflict,
					Message:    err.Err.Error(),
					Code:       "CONFLICT",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *common.BadRequestError:
				errorResponse = common.ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    util.ExtractErrorPart(err.Err.Error()),
					Code:       "BAD_REQUEST",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			default:
				errorResponse = handlePossibleDockerError(ctx, err.Err)
			}

			if errorResponse.StatusCode == http.StatusInternalServerError {
				log.WithError(err).WithFields(log.Fields{
					"path":   ctx.Request.URL.Path,
					"method": ctx.Request.Method,
				}).Error("Internal Server Error")
			} else {
				log.WithFields(log.Fields{
					"method": ctx.Request.Method,
					"URI":    ctx.Request.URL.Path,
					"status": errorResponse.StatusCode,
					"error":  errorResponse.Message,
				}).Error("API ERROR")
			}

			// Set explicit content type header
			ctx.Header("Content-Type", "application/json")
			ctx.JSON(errorResponse.StatusCode, errorResponse)
		}
	}
}

func handlePossibleDockerError(ctx *gin.Context, err error) common.ErrorResponse {
	if errdefs.IsNotFound(err) {
		return common.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("resource not found: %s", err.Error()),
			Code:       "NOT_FOUND",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsUnauthorized(err) || strings.Contains(err.Error(), "unauthorized") {
		return common.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    fmt.Sprintf("unauthorized: %s", err.Error()),
			Code:       "UNAUTHORIZED",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsConflict(err) {
		return common.ErrorResponse{
			StatusCode: http.StatusConflict,
			Message:    fmt.Sprintf("conflict: %s", err.Error()),
			Code:       "CONFLICT",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsInvalidParameter(err) {
		return common.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("bad request: %s", err.Error()),
			Code:       "BAD_REQUEST",
			Timestamp:  time.Now(),
			Path:       ctx.Request.URL.Path,
			Method:     ctx.Request.Method,
		}
	} else if errdefs.IsSystem(err) {
		if strings.Contains(err.Error(), "unable to find user") {
			return common.ErrorResponse{
				StatusCode: http.StatusBadRequest,
				Message:    util.ExtractErrorPart(err.Error()),
				Code:       "BAD_REQUEST",
				Timestamp:  time.Now(),
				Path:       ctx.Request.URL.Path,
				Method:     ctx.Request.Method,
			}
		}
	}

	return common.ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Code:       "INTERNAL_SERVER_ERROR",
		Timestamp:  time.Now(),
		Path:       ctx.Request.URL.Path,
		Method:     ctx.Request.Method,
	}
}
