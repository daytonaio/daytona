// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
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
			case *ForbiddenError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusForbidden,
					Message:    err.Err.Error(),
					Code:       "FORBIDDEN",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *RequestTimeoutError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusRequestTimeout,
					Message:    err.Err.Error(),
					Code:       "REQUEST_TIMEOUT",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *GoneError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusGone,
					Message:    err.Err.Error(),
					Code:       "GONE",
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}
			case *InternalServerError:
				errorResponse = ErrorResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Err.Error(),
					Code:       "INTERNAL_SERVER_ERROR",
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
					"error":  errorResponse.Message,
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

const maxStackTraceSize = 64 * 1024 // 64 KB

func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if errType, ok := err.(error); ok && errors.Is(errType, http.ErrAbortHandler) {
					// Do nothing, the request was aborted
					return
				}

				log.Errorf("panic recovered: %v", err)
				// print caller stack
				buf := make([]byte, maxStackTraceSize)
				stackSize := runtime.Stack(buf, false)
				log.Errorf("stack trace: %s", string(buf[:stackSize]))

				if ctx.Writer.Written() {
					return
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}
