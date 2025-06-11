// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func NewErrorMiddleware(defaultErrorHandler func(ctx *gin.Context, err error) ErrorResponse) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		errs := ctx.Errors
		if len(errs) > 0 {
			var errorResponse ErrorResponse
			err := errs.Last()

			switch e := err.Err.(type) {
			case *CustomError:
				errorResponse = ErrorResponse{
					StatusCode: e.StatusCode,
					Message:    e.Message,
					Code:       e.Code,
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *NotFoundError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusNotFound,
					Message:    err.Err.Error(),
					Code:       "NOT_FOUND",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *UnauthorizedError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    err.Err.Error(),
					Code:       "UNAUTHORIZED",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *InvalidBodyRequestError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    err.Err.Error(),
					Code:       "INVALID_REQUEST_BODY",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *ConflictError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusConflict,
					Message:    err.Err.Error(),
					Code:       "CONFLICT",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *BadRequestError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    ExtractErrorPart(err.Err.Error()),
					Code:       "BAD_REQUEST",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			default:
				errorResponse = defaultErrorHandler(ctx, err)
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

func ExtractErrorPart(errorMsg string) string {
	r := regexp.MustCompile(`(unable to find user [^:]+)`)

	matches := r.FindStringSubmatch(errorMsg)

	if len(matches) < 2 {
		return errorMsg
	}

	return fmt.Sprintf("bad request: %s", matches[1])
}
