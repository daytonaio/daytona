// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// DaytonaError is the base error type for all Daytona SDK errors
type DaytonaError struct {
	Message    string
	StatusCode int
	Headers    http.Header
}

func (e *DaytonaError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("Daytona error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Daytona error: %s", e.Message)
}

// NewDaytonaError creates a new DaytonaError
func NewDaytonaError(message string, statusCode int, headers http.Header) *DaytonaError {
	return &DaytonaError{
		Message:    message,
		StatusCode: statusCode,
		Headers:    headers,
	}
}

// DaytonaBadRequestError represents a malformed request or invalid parameters (400).
// Use errors.As to check for this type.
type DaytonaBadRequestError struct {
	*DaytonaError
}

func (e *DaytonaBadRequestError) Error() string {
	return fmt.Sprintf("Bad request: %s", e.Message)
}

// NewDaytonaBadRequestError creates a new DaytonaBadRequestError
func NewDaytonaBadRequestError(message string, headers http.Header) *DaytonaBadRequestError {
	return &DaytonaBadRequestError{
		DaytonaError: NewDaytonaError(message, http.StatusBadRequest, headers),
	}
}

// DaytonaAuthenticationError represents an authentication failure (401).
// Use errors.As to check for this type.
type DaytonaAuthenticationError struct {
	*DaytonaError
}

func (e *DaytonaAuthenticationError) Error() string {
	return fmt.Sprintf("Authentication failed: %s", e.Message)
}

// NewDaytonaAuthenticationError creates a new DaytonaAuthenticationError
func NewDaytonaAuthenticationError(message string, headers http.Header) *DaytonaAuthenticationError {
	return &DaytonaAuthenticationError{
		DaytonaError: NewDaytonaError(message, http.StatusUnauthorized, headers),
	}
}

// DaytonaForbiddenError represents an authorization failure (403).
// Use errors.As to check for this type.
type DaytonaForbiddenError struct {
	*DaytonaError
}

func (e *DaytonaForbiddenError) Error() string {
	return fmt.Sprintf("Forbidden: %s", e.Message)
}

// NewDaytonaForbiddenError creates a new DaytonaForbiddenError
func NewDaytonaForbiddenError(message string, headers http.Header) *DaytonaForbiddenError {
	return &DaytonaForbiddenError{
		DaytonaError: NewDaytonaError(message, http.StatusForbidden, headers),
	}
}

// DaytonaNotFoundError represents a resource not found error (404).
// Use errors.As to check for this type.
type DaytonaNotFoundError struct {
	*DaytonaError
}

func (e *DaytonaNotFoundError) Error() string {
	return fmt.Sprintf("Resource not found: %s", e.Message)
}

// NewDaytonaNotFoundError creates a new DaytonaNotFoundError
func NewDaytonaNotFoundError(message string, headers http.Header) *DaytonaNotFoundError {
	return &DaytonaNotFoundError{
		DaytonaError: NewDaytonaError(message, http.StatusNotFound, headers),
	}
}

// DaytonaConflictError represents a resource conflict (409).
// Use errors.As to check for this type.
type DaytonaConflictError struct {
	*DaytonaError
}

func (e *DaytonaConflictError) Error() string {
	return fmt.Sprintf("Conflict: %s", e.Message)
}

// NewDaytonaConflictError creates a new DaytonaConflictError
func NewDaytonaConflictError(message string, headers http.Header) *DaytonaConflictError {
	return &DaytonaConflictError{
		DaytonaError: NewDaytonaError(message, http.StatusConflict, headers),
	}
}

// DaytonaValidationError represents a semantic validation failure (422).
// Use errors.As to check for this type.
type DaytonaValidationError struct {
	*DaytonaError
}

func (e *DaytonaValidationError) Error() string {
	return fmt.Sprintf("Validation error: %s", e.Message)
}

// NewDaytonaValidationError creates a new DaytonaValidationError
func NewDaytonaValidationError(message string, headers http.Header) *DaytonaValidationError {
	return &DaytonaValidationError{
		DaytonaError: NewDaytonaError(message, http.StatusUnprocessableEntity, headers),
	}
}

// DaytonaRateLimitError represents a rate limit error (429).
// Use errors.As to check for this type.
type DaytonaRateLimitError struct {
	*DaytonaError
}

func (e *DaytonaRateLimitError) Error() string {
	return fmt.Sprintf("Rate limit exceeded: %s", e.Message)
}

// NewDaytonaRateLimitError creates a new DaytonaRateLimitError
func NewDaytonaRateLimitError(message string, headers http.Header) *DaytonaRateLimitError {
	return &DaytonaRateLimitError{
		DaytonaError: NewDaytonaError(message, http.StatusTooManyRequests, headers),
	}
}

// DaytonaServerError represents an unexpected server-side failure (5xx).
// Use errors.As to check for this type.
type DaytonaServerError struct {
	*DaytonaError
}

func (e *DaytonaServerError) Error() string {
	return fmt.Sprintf("Server error (status %d): %s", e.StatusCode, e.Message)
}

// NewDaytonaServerError creates a new DaytonaServerError
func NewDaytonaServerError(message string, statusCode int, headers http.Header) *DaytonaServerError {
	return &DaytonaServerError{
		DaytonaError: NewDaytonaError(message, statusCode, headers),
	}
}

// DaytonaTimeoutError represents a timeout error.
// Use errors.As to check for this type.
type DaytonaTimeoutError struct {
	*DaytonaError
}

func (e *DaytonaTimeoutError) Error() string {
	return fmt.Sprintf("Operation timed out: %s", e.Message)
}

// NewDaytonaTimeoutError creates a new DaytonaTimeoutError
func NewDaytonaTimeoutError(message string) *DaytonaTimeoutError {
	return &DaytonaTimeoutError{
		DaytonaError: NewDaytonaError(message, 0, nil),
	}
}

// DaytonaConnectionError represents a network-level connection failure (no HTTP response).
// Use errors.As to check for this type.
type DaytonaConnectionError struct {
	*DaytonaError
}

func (e *DaytonaConnectionError) Error() string {
	return fmt.Sprintf("Connection error: %s", e.Message)
}

// NewDaytonaConnectionError creates a new DaytonaConnectionError
func NewDaytonaConnectionError(message string) *DaytonaConnectionError {
	return &DaytonaConnectionError{
		DaytonaError: NewDaytonaError(message, 0, nil),
	}
}

// mapStatusCode maps an HTTP status code to the appropriate DaytonaError subtype
func mapStatusCode(message string, statusCode int, headers http.Header) error {
	switch statusCode {
	case http.StatusBadRequest:
		return NewDaytonaBadRequestError(message, headers)
	case http.StatusUnauthorized:
		return NewDaytonaAuthenticationError(message, headers)
	case http.StatusForbidden:
		return NewDaytonaForbiddenError(message, headers)
	case http.StatusNotFound:
		return NewDaytonaNotFoundError(message, headers)
	case http.StatusConflict:
		return NewDaytonaConflictError(message, headers)
	case http.StatusUnprocessableEntity:
		return NewDaytonaValidationError(message, headers)
	case http.StatusTooManyRequests:
		return NewDaytonaRateLimitError(message, headers)
	case 0:
		// Network or client error (no HTTP response)
		return NewDaytonaConnectionError(message)
	default:
		if statusCode >= 500 {
			return NewDaytonaServerError(message, statusCode, headers)
		}
		return NewDaytonaError(message, statusCode, headers)
	}
}

// extractErrorMessage attempts to extract a human-readable message from a raw error body
func extractErrorMessage(body []byte, fallback string) string {
	if len(body) == 0 {
		return fallback
	}

	var errResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil {
		if errResp.Message != "" {
			return errResp.Message
		}
		if errResp.Error != "" {
			return errResp.Error
		}
	}

	return string(body)
}

// ConvertAPIError converts api-client-go errors to SDK error types
func ConvertAPIError(err error, httpResp *http.Response) error {
	if err == nil {
		return nil
	}

	var message string
	var statusCode int
	var headers http.Header

	if httpResp != nil {
		statusCode = httpResp.StatusCode
		headers = httpResp.Header
	}

	if genErr, ok := err.(*apiclient.GenericOpenAPIError); ok {
		message = extractErrorMessage(genErr.Body(), genErr.Error())
	} else {
		message = err.Error()
	}

	return mapStatusCode(message, statusCode, headers)
}

// ConvertToolboxError converts toolbox-api-client-go errors to SDK error types
func ConvertToolboxError(err error, httpResp *http.Response) error {
	if err == nil {
		return nil
	}

	var message string
	var statusCode int
	var headers http.Header

	if httpResp != nil {
		statusCode = httpResp.StatusCode
		headers = httpResp.Header
	}

	if genErr, ok := err.(*toolbox.GenericOpenAPIError); ok {
		message = extractErrorMessage(genErr.Body(), genErr.Error())
	} else {
		message = err.Error()
	}

	return mapStatusCode(message, statusCode, headers)
}
