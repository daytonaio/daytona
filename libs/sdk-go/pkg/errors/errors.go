// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/toolbox-api-go"
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

// DaytonaNotFoundError represents a resource not found error (404)
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

// DaytonaRateLimitError represents a rate limit error (429)
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

// DaytonaTimeoutError represents a timeout error
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

	// Try to extract message from GenericOpenAPIError
	if genErr, ok := err.(*apiclient.GenericOpenAPIError); ok {
		body := genErr.Body()
		if len(body) > 0 {
			// Try to parse as JSON
			var errResp struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if json.Unmarshal(body, &errResp) == nil {
				if errResp.Message != "" {
					message = errResp.Message
				} else if errResp.Error != "" {
					message = errResp.Error
				}
			}

			// Fall back to raw body if no structured message
			if message == "" {
				message = string(body)
			}
		}

		// Fall back to error string if no body
		if message == "" {
			message = genErr.Error()
		}
	} else {
		message = err.Error()
	}

	// Map status codes to SDK error types
	switch statusCode {
	case http.StatusNotFound:
		return NewDaytonaNotFoundError(message, headers)
	case http.StatusTooManyRequests:
		return NewDaytonaRateLimitError(message, headers)
	case 0:
		// Network or client error (no HTTP response)
		return NewDaytonaError(message, 0, nil)
	default:
		return NewDaytonaError(message, statusCode, headers)
	}
}

// ConvertToolboxError converts toolbox-api-go errors to SDK error types
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

	// Try to extract message from GenericOpenAPIError
	if genErr, ok := err.(*toolbox.GenericOpenAPIError); ok {
		body := genErr.Body()
		if len(body) > 0 {
			// Try to parse as JSON
			var errResp struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if json.Unmarshal(body, &errResp) == nil {
				if errResp.Message != "" {
					message = errResp.Message
				} else if errResp.Error != "" {
					message = errResp.Error
				}
			}

			// Fall back to raw body if no structured message
			if message == "" {
				message = string(body)
			}
		}

		// Fall back to error string if no body
		if message == "" {
			message = genErr.Error()
		}
	} else {
		message = err.Error()
	}

	// Map status codes to SDK error types
	switch statusCode {
	case http.StatusNotFound:
		return NewDaytonaNotFoundError(message, headers)
	case http.StatusTooManyRequests:
		return NewDaytonaRateLimitError(message, headers)
	case 0:
		// Network or client error (no HTTP response)
		return NewDaytonaError(message, 0, nil)
	default:
		return NewDaytonaError(message, statusCode, headers)
	}
}
