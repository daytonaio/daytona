// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// DefaultInternalServerErrorHandler always returns a 500 envelope with the
// raw err.Error(). Use as the defaultErrorHandler when a service has no
// custom classification (recording dashboard, tests).
func DefaultInternalServerErrorHandler(source string) func(*gin.Context, error) ErrorResponse {
	return func(ctx *gin.Context, err error) ErrorResponse {
		return NewErrorResponseForCtx(ctx, http.StatusInternalServerError, source, err.Error())
	}
}

func NewErrorMiddleware(source string, defaultErrorHandler func(ctx *gin.Context, err error) ErrorResponse) gin.HandlerFunc {
	if defaultErrorHandler == nil {
		defaultErrorHandler = DefaultInternalServerErrorHandler(source)
	}

	return func(ctx *gin.Context) {
		ctx.Next()

		// Do not override the response if it has already been written
		if ctx.Writer.Written() {
			return
		}

		errs := ctx.Errors
		if len(errs) > 0 {
			var errorResponse ErrorResponse
			// Unwrap *gin.Error once so the default handler never has to.
			underlying := errs.Last().Err

			// Self-describing errors go straight to the wire.
			if he, ok := underlying.(HTTPError); ok {
				errorResponse = NewErrorResponseFromHTTPError(ctx, source, he)
			} else {
				errorResponse = defaultErrorHandler(ctx, underlying)
			}

			if errorResponse.StatusCode == http.StatusInternalServerError {
				slog.Error("Internal Server Error",
					"path", ctx.Request.URL.Path,
					"method", ctx.Request.Method,
					"error", errorResponse.Message,
				)
			} else {
				slog.Error("API ERROR",
					"method", ctx.Request.Method,
					"URI", ctx.Request.URL.Path,
					"status", errorResponse.StatusCode,
					"error", errorResponse.Message,
				)
			}

			// Set explicit content type header
			ctx.Header("Content-Type", "application/json")
			ctx.JSON(errorResponse.StatusCode, errorResponse)
		}
	}
}

// NewErrorResponseForCtx builds an ErrorResponse with path/method/timestamp
// populated from ctx. For default error handlers.
func NewErrorResponseForCtx(ctx *gin.Context, statusCode int, source, message string) ErrorResponse {
	return ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Source:     source,
		Timestamp:  time.Now(),
		Path:       ctx.Request.URL.Path,
		Method:     ctx.Request.Method,
	}
}

// NewErrorResponseFromHTTPError builds an ErrorResponse from a self-describing
// HTTPError. BadRequestError messages are rewritten by ExtractErrorPart.
func NewErrorResponseFromHTTPError(ctx *gin.Context, source string, he HTTPError) ErrorResponse {
	msg := he.Error()
	if _, isBadReq := he.(*BadRequestError); isBadReq {
		msg = ExtractErrorPart(msg)
	}
	resp := NewErrorResponseForCtx(ctx, he.HTTPStatusCode(), source, msg)
	resp.Code = he.ErrorCode()
	return resp
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

				slog.Error("panic recovered", "panic", err)
				// print caller stack
				buf := make([]byte, maxStackTraceSize)
				stackSize := runtime.Stack(buf, false)
				slog.Error("stack trace", "stack", string(buf[:stackSize]))

				if ctx.Writer.Written() {
					return
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}
