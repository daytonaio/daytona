// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package clierr defines the CLI error model: categorized errors with an
// optional remediation hint and a deterministic process exit code mapping.
// It is a leaf package and must only import the standard library.
package clierr

import (
	"errors"
	"fmt"
)

// Category classifies an error so callers (and scripts parsing --format
// output) can react to the kind of failure without string matching.
type Category string

const (
	CategoryUsage     Category = "usage"
	CategoryAuth      Category = "auth"
	CategoryNotFound  Category = "not_found"
	CategoryConflict  Category = "conflict"
	CategoryRateLimit Category = "rate_limit"
	CategoryServer    Category = "server"
	CategoryNetwork   Category = "network"
	CategoryTimeout   Category = "timeout"
)

// Error is the structured CLI error. Message holds the failure description,
// Hint an optional remediation suggestion, and Code an optional explicit
// process exit code overriding the category mapping.
type Error struct {
	Category Category
	Message  string
	Hint     string
	Code     int
}

func (e *Error) Error() string {
	if e.Hint != "" {
		return e.Message + " - " + e.Hint
	}
	return e.Message
}

func New(cat Category, msg string) *Error {
	return &Error{Category: cat, Message: msg}
}

func Newf(cat Category, format string, a ...any) *Error {
	return &Error{Category: cat, Message: fmt.Sprintf(format, a...)}
}

// WithHint sets the remediation hint and returns the error for chaining.
func (e *Error) WithHint(hint string) *Error {
	e.Hint = hint
	return e
}

// WithCode sets an explicit process exit code and returns the error for chaining.
func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

// HasCategory reports whether err is (or wraps) an *Error with the given category.
func HasCategory(err error, cat Category) bool {
	var cliErr *Error
	return errors.As(err, &cliErr) && cliErr.Category == cat
}

// ExitCode maps an error to a process exit code: an explicit Code wins,
// usage errors exit 2, wait timeouts exit 124, everything else exits 1.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var cliErr *Error
	if !errors.As(err, &cliErr) {
		return 1
	}

	if cliErr.Code > 0 {
		return cliErr.Code
	}

	switch cliErr.Category {
	case CategoryUsage:
		return 2
	case CategoryTimeout:
		return 124
	default:
		return 1
	}
}

// FromHTTPStatus builds an Error categorized by the HTTP response status,
// attaching the standard remediation hint for authentication failures.
func FromHTTPStatus(status int, msg string) *Error {
	switch {
	case status == 400:
		return New(CategoryUsage, msg)
	case status == 401:
		return New(CategoryAuth, msg).WithHint("run 'daytona login' to reauthenticate")
	case status == 403:
		return New(CategoryAuth, msg).WithHint("check that your API key has sufficient permissions for this action")
	case status == 404:
		return New(CategoryNotFound, msg)
	case status == 409:
		return New(CategoryConflict, msg)
	case status == 429:
		return New(CategoryRateLimit, msg)
	case status >= 500:
		return New(CategoryServer, msg)
	default:
		return New(CategoryServer, msg)
	}
}
