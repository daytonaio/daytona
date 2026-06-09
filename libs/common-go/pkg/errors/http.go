// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorResponse is the wire shape of every error response emitted by a Daytona
// service. (source, code) is the canonical machine-readable identifier; code
// is optional and only set when the error needs to be distinguished beyond
// its HTTP status.
//
//	@Description	Error response
//	@Schema			ErrorResponse
type ErrorResponse struct {
	StatusCode int       `json:"statusCode" example:"400" binding:"required"`
	Message    string    `json:"message" example:"Bad request" binding:"required"`
	Source     string    `json:"source,omitempty" example:"DAYTONA_DAEMON"`
	Code       string    `json:"code,omitempty" example:"BAD_REQUEST"`
	Timestamp  time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z" binding:"required"`
	Path       string    `json:"path" example:"/api/resource" binding:"required"`
	Method     string    `json:"method,omitempty" example:"GET"`
} //	@name	ErrorResponse

// HTTPError is an error that self-describes its HTTP status (and optional
// machine code). The middleware emits such errors directly onto the wire.
type HTTPError interface {
	error
	HTTPStatusCode() int
	ErrorCode() string // "" when no machine code applies
}

// CustomError is the generic typed error: arbitrary status + machine code.
type CustomError struct {
	StatusCode int
	Message    string
	Code       string
}

func (e *CustomError) IsRetryable() bool {
	switch e.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}

	return false
}

func (e *CustomError) Error() string       { return e.Message }
func (e *CustomError) HTTPStatusCode() int { return e.StatusCode }
func (e *CustomError) ErrorCode() string   { return e.Code }

// NewCustomError builds a CustomError; use 502/503/504 for retryable cases.
func NewCustomError(statusCode int, message, code string) error {
	return &CustomError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
	}
}

// Per-status typed errors below. Each carries a canned status + canned code
// (NOT_FOUND, CONFLICT, ...) preserved from main for backward compatibility
// with existing SDK consumers.

type NotFoundError struct{ Message string }

func NewNotFoundError(err error) error {
	return &NotFoundError{Message: fmt.Sprintf("not found: %s", err.Error())}
}
func (e *NotFoundError) Error() string       { return e.Message }
func (e *NotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *NotFoundError) ErrorCode() string   { return "NOT_FOUND" }

func IsNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

type InvalidBodyRequestError struct{ Message string }

func NewInvalidBodyRequestError(err error) error {
	return &InvalidBodyRequestError{Message: fmt.Sprintf("invalid body request: %s", err.Error())}
}
func (e *InvalidBodyRequestError) Error() string       { return e.Message }
func (e *InvalidBodyRequestError) HTTPStatusCode() int { return http.StatusBadRequest }
func (e *InvalidBodyRequestError) ErrorCode() string   { return "INVALID_REQUEST_BODY" }

func IsInvalidBodyRequestError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "invalid body request")
}

type UnauthorizedError struct{ Message string }

func NewUnauthorizedError(err error) error {
	return &UnauthorizedError{Message: fmt.Sprintf("unauthorized: %s", err.Error())}
}
func (e *UnauthorizedError) Error() string       { return e.Message }
func (e *UnauthorizedError) HTTPStatusCode() int { return http.StatusUnauthorized }
func (e *UnauthorizedError) ErrorCode() string   { return "UNAUTHORIZED" }

func IsUnauthorizedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unauthorized")
}

type ConflictError struct{ Message string }

func NewConflictError(err error) error {
	return &ConflictError{Message: fmt.Sprintf("conflict: %s", err.Error())}
}
func (e *ConflictError) Error() string       { return e.Message }
func (e *ConflictError) HTTPStatusCode() int { return http.StatusConflict }
func (e *ConflictError) ErrorCode() string   { return "CONFLICT" }

func IsConflictError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "conflict")
}

// BadRequestError's message is rewritten by ExtractErrorPart in the middleware.
type BadRequestError struct{ Message string }

func NewBadRequestError(err error) error {
	return &BadRequestError{Message: fmt.Sprintf("bad request: %s", err.Error())}
}
func (e *BadRequestError) Error() string       { return e.Message }
func (e *BadRequestError) HTTPStatusCode() int { return http.StatusBadRequest }
func (e *BadRequestError) ErrorCode() string   { return "BAD_REQUEST" }

func IsBadRequestError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "bad request")
}

type ForbiddenError struct{ Message string }

func NewForbiddenError(err error) error {
	return &ForbiddenError{Message: fmt.Sprintf("forbidden: %s", err.Error())}
}
func (e *ForbiddenError) Error() string       { return e.Message }
func (e *ForbiddenError) HTTPStatusCode() int { return http.StatusForbidden }
func (e *ForbiddenError) ErrorCode() string   { return "FORBIDDEN" }

func IsForbiddenError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "forbidden")
}

type RequestTimeoutError struct{ Message string }

func NewRequestTimeoutError(err error) error {
	return &RequestTimeoutError{Message: fmt.Sprintf("request timeout: %s", err.Error())}
}
func (e *RequestTimeoutError) Error() string       { return e.Message }
func (e *RequestTimeoutError) HTTPStatusCode() int { return http.StatusRequestTimeout }
func (e *RequestTimeoutError) ErrorCode() string   { return "REQUEST_TIMEOUT" }

func IsRequestTimeoutError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "request timeout")
}

type GoneError struct{ Message string }

func NewGoneError(err error) error {
	return &GoneError{Message: fmt.Sprintf("gone: %s", err.Error())}
}
func (e *GoneError) Error() string       { return e.Message }
func (e *GoneError) HTTPStatusCode() int { return http.StatusGone }
func (e *GoneError) ErrorCode() string   { return "GONE" }

func IsGoneError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "gone")
}

type InternalServerError struct{ Message string }

func NewInternalServerError(err error) error {
	return &InternalServerError{Message: fmt.Sprintf("internal server error: %s", err.Error())}
}
func (e *InternalServerError) Error() string       { return e.Message }
func (e *InternalServerError) HTTPStatusCode() int { return http.StatusInternalServerError }
func (e *InternalServerError) ErrorCode() string   { return "INTERNAL_SERVER_ERROR" }

func IsInternalServerError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "internal server error")
}

type UnprocessableEntityError struct{ Message string }

func NewUnprocessableEntityError(err error) error {
	return &UnprocessableEntityError{Message: fmt.Sprintf("unprocessable entity: %s", err.Error())}
}
func (e *UnprocessableEntityError) Error() string       { return e.Message }
func (e *UnprocessableEntityError) HTTPStatusCode() int { return http.StatusUnprocessableEntity }
func (e *UnprocessableEntityError) ErrorCode() string   { return "UNPROCESSABLE_ENTITY" }
