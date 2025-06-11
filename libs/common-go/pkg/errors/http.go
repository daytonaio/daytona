// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorResponse represents the error response structure
//
//	@Description	Error response
//	@Schema			ErrorResponse
type ErrorResponse struct {
	StatusCode int       `json:"statusCode" example:"400" binding:"required"`
	Message    string    `json:"message" example:"Bad request" binding:"required"`
	Code       string    `json:"code" example:"BAD_REQUEST" binding:"required"`
	Timestamp  time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z" binding:"required"`
	Path       string    `json:"path" example:"/api/resource" binding:"required"`
	Method     string    `json:"method" example:"GET" binding:"required"`
} //	@name	ErrorResponse

type CustomError struct {
	StatusCode int
	Message    string
	Code       string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(statusCode int, message, code string) error {
	return &CustomError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
	}
}

type NotFoundError struct {
	Message string
}

func NewNotFoundError(err error) error {
	return &NotFoundError{
		Message: fmt.Sprintf("not found: %s", err.Error()),
	}
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func IsNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

type InvalidBodyRequestError struct {
	Message string
}

func (e *InvalidBodyRequestError) Error() string {
	return e.Message
}

func NewInvalidBodyRequestError(err error) error {
	return &InvalidBodyRequestError{
		Message: fmt.Sprintf("invalid body request: %s", err.Error()),
	}
}

func IsInvalidBodyRequestError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "invalid body request")
}

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

func NewUnauthorizedError(err error) error {
	return &UnauthorizedError{
		Message: fmt.Sprintf("unauthorized: %s", err.Error()),
	}
}

func IsUnauthorizedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unauthorized")
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

func NewConflictError(err error) error {
	return &ConflictError{
		Message: fmt.Sprintf("conflict: %s", err.Error()),
	}
}

func IsConflictError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "conflict")
}

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

func NewBadRequestError(err error) error {
	return &BadRequestError{
		Message: fmt.Sprintf("bad request: %s", err.Error()),
	}
}

func IsBadRequestError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "bad request")
}
