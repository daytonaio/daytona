// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Package errors defines the typed error model used by the Daytona Go SDK.
//
// Every error returned by the SDK is a `*DaytonaError` carrying a
// human-readable message, the HTTP `StatusCode` (when applicable), an
// optional machine-readable `Code` / `Source` pair, and the response
// `Headers`. There are no per-status struct types — a single concrete type
// keeps the surface small and unambiguous.
//
// Branching is done with `errors.Is` against the package-level sentinels:
//
//	if errors.Is(err, sdkerrors.ErrNotFound) {
//	    // any HTTP 404 from any source
//	}
//	if errors.Is(err, sdkerrors.ErrGitAuthFailed) {
//	    // precisely DAYTONA_DAEMON / GIT_AUTH_FAILED
//	}
//	if errors.Is(err, sdkerrors.ErrAuthentication) {
//	    // the same git-auth error ALSO matches the broader 401 sentinel,
//	    // mirroring the inheritance hierarchy of the other Daytona SDKs.
//	}
//
// Reading metadata off an error is done with `errors.As`:
//
//	var de *sdkerrors.DaytonaError
//	if errors.As(err, &de) {
//	    log.Printf("status=%d code=%s source=%s", de.StatusCode, de.Code, de.Source)
//	}
package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	toolbox "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// ----- Source identifiers (wire-format, shared with the other SDKs) -----
//
// Set by the translation layer when a Daytona service stamps them on the
// wire envelope. An empty `Source` means the response did not carry a
// structured envelope (treat as opaque).

const (
	SourceAPI    = "DAYTONA_API"
	SourceDaemon = "DAYTONA_DAEMON"
	SourceProxy  = "DAYTONA_PROXY"
)

// ----- DaytonaError: the one and only concrete error type -----

// DaytonaError is the single error type returned by the SDK. Use
// `errors.As(err, &target *DaytonaError)` to read its fields and
// `errors.Is(err, sentinel)` to branch on the kind.
type DaytonaError struct {
	Message    string
	StatusCode int
	Code       string
	Source     string
	Headers    http.Header
}

func (e *DaytonaError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("Daytona error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Daytona error: %s", e.Message)
}

// Is implements the `errors.Is` contract. A target matches when it is one
// of the package-level sentinels and either:
//
//   - the target carries a non-empty `Code`, in which case BOTH `Source`
//     and `Code` must match exactly (domain-code sentinel), or
//   - the target carries a non-zero `StatusCode`, in which case the
//     receiver's `StatusCode` must match (status-class sentinel).
//
// Because the SDK always stamps the HTTP status alongside the domain code,
// `errors.Is(err, ErrGitAuthFailed)` and `errors.Is(err, ErrAuthentication)`
// both match the same underlying error — mirroring the inheritance
// hierarchy used by the Python/TypeScript/Java SDKs.
func (e *DaytonaError) Is(target error) bool {
	t, ok := target.(*DaytonaError)
	if !ok {
		return false
	}
	if t.Code != "" {
		return e.Code == t.Code && e.Source == t.Source
	}
	if t.StatusCode != 0 && e.StatusCode == t.StatusCode {
		return true
	}
	return false
}

// NewDaytonaError builds a DaytonaError with the given message, status code
// and headers. `Source` is left empty — set it explicitly via `SourceSDK`
// for SDK-internal errors, or via the translation layer for server-side
// errors. Most callers should use this directly; the sentinels below are
// for branching with `errors.Is`, not for constructing errors.
func NewDaytonaError(message string, statusCode int, headers http.Header) *DaytonaError {
	return &DaytonaError{Message: message, StatusCode: statusCode, Headers: headers}
}

// NewDaytonaTimeoutError is a convenience constructor for client-side
// timeouts. Equivalent to `NewDaytonaError(message, http.StatusRequestTimeout, nil)`.
func NewDaytonaTimeoutError(message string) *DaytonaError {
	return NewDaytonaError(message, http.StatusRequestTimeout, nil)
}

// NewDaytonaConnectionError is a convenience constructor for transport-level
// failures with no HTTP response (DNS, dial, TLS, mid-request drop).
func NewDaytonaConnectionError(message string) *DaytonaError {
	return NewDaytonaError(message, 0, nil)
}

// ----- Sentinels -----

var (
	// HTTP status-class sentinels. Names follow HTTP terminology.
	ErrBadRequest          = &DaytonaError{StatusCode: http.StatusBadRequest}
	ErrAuthentication      = &DaytonaError{StatusCode: http.StatusUnauthorized}
	ErrForbidden           = &DaytonaError{StatusCode: http.StatusForbidden}
	ErrNotFound            = &DaytonaError{StatusCode: http.StatusNotFound}
	ErrTimeout             = &DaytonaError{StatusCode: http.StatusRequestTimeout}
	ErrConflict            = &DaytonaError{StatusCode: http.StatusConflict}
	ErrGone                = &DaytonaError{StatusCode: http.StatusGone}
	ErrUnprocessableEntity = &DaytonaError{StatusCode: http.StatusUnprocessableEntity}
	ErrRateLimit           = &DaytonaError{StatusCode: http.StatusTooManyRequests}
	ErrInternalServer      = &DaytonaError{StatusCode: http.StatusInternalServerError}
	ErrBadGateway          = &DaytonaError{StatusCode: http.StatusBadGateway}
	ErrServiceUnavailable  = &DaytonaError{StatusCode: http.StatusServiceUnavailable}

	// Deprecated: use ErrBadRequest. Kept so existing callers do not break.
	ErrValidation = ErrBadRequest
	// Deprecated: use ErrForbidden. Kept so existing callers do not break.
	ErrAuthorization = ErrForbidden

	// Daemon: git.
	ErrGitAuthFailed     = &DaytonaError{Source: SourceDaemon, Code: "GIT_AUTH_FAILED"}
	ErrGitRepoNotFound   = &DaytonaError{Source: SourceDaemon, Code: "GIT_REPO_NOT_FOUND"}
	ErrGitBranchNotFound = &DaytonaError{Source: SourceDaemon, Code: "GIT_BRANCH_NOT_FOUND"}
	ErrGitBranchExists   = &DaytonaError{Source: SourceDaemon, Code: "GIT_BRANCH_EXISTS"}
	ErrGitPushRejected   = &DaytonaError{Source: SourceDaemon, Code: "GIT_PUSH_REJECTED"}
	ErrGitDirtyWorktree  = &DaytonaError{Source: SourceDaemon, Code: "GIT_DIRTY_WORKTREE"}
	ErrGitMergeConflict  = &DaytonaError{Source: SourceDaemon, Code: "GIT_MERGE_CONFLICT"}

	// Daemon: filesystem.
	ErrFileNotFound     = &DaytonaError{Source: SourceDaemon, Code: "FILE_NOT_FOUND"}
	ErrFileAccessDenied = &DaytonaError{Source: SourceDaemon, Code: "FILE_ACCESS_DENIED"}

	// Daemon: LSP.
	ErrLspServerNotInitialized = &DaytonaError{Source: SourceDaemon, Code: "LSP_SERVER_NOT_INITIALIZED"}

	// Daemon: process / session.
	ErrProcessExecutionTimeout = &DaytonaError{Source: SourceDaemon, Code: "PROCESS_EXECUTION_TIMEOUT"}
	ErrProcessNotFound         = &DaytonaError{Source: SourceDaemon, Code: "PROCESS_NOT_FOUND"}
	ErrSessionEnded            = &DaytonaError{Source: SourceDaemon, Code: "SESSION_ENDED"}
	ErrCommandAlreadyCompleted = &DaytonaError{Source: SourceDaemon, Code: "COMMAND_ALREADY_COMPLETED"}

	// Daemon: computer-use.
	ErrA11yUnavailable         = &DaytonaError{Source: SourceDaemon, Code: "A11Y_UNAVAILABLE"}
	ErrRecordingStillActive    = &DaytonaError{Source: SourceDaemon, Code: "RECORDING_STILL_ACTIVE"}
	ErrRecordingFfmpegNotFound = &DaytonaError{Source: SourceDaemon, Code: "RECORDING_FFMPEG_NOT_FOUND"}
)

// ----- Wire-envelope parsing and conversion from generated clients -----

func parseErrorBody(body []byte) (message, code, source string, parsedStatusCode int) {
	if len(body) == 0 {
		return "", "", "", 0
	}

	var errResp struct {
		Message    string `json:"message"`
		Error      string `json:"error"`
		StatusCode int    `json:"statusCode"`
		Code       string `json:"code"`
		Source     string `json:"source"`
	}

	if json.Unmarshal(body, &errResp) != nil {
		return string(body), "", "", 0
	}

	if errResp.Message != "" {
		message = errResp.Message
	} else if errResp.Error != "" {
		message = errResp.Error
	}

	return message, errResp.Code, errResp.Source, errResp.StatusCode
}

// NewDaytonaErrorFromBody parses a JSON response body and builds a
// DaytonaError. When the body carries its own `statusCode` field that
// overrides the caller-supplied one (server-side envelopes are authoritative).
func NewDaytonaErrorFromBody(body []byte, statusCode int, headers http.Header) *DaytonaError {
	message, code, source, parsedStatusCode := parseErrorBody(body)
	if parsedStatusCode != 0 {
		statusCode = parsedStatusCode
	}
	if message == "" {
		message = "Request failed"
	}
	return &DaytonaError{
		Message:    message,
		StatusCode: statusCode,
		Code:       code,
		Source:     source,
		Headers:    headers,
	}
}

// ConvertAPIError converts an error returned by the generated api-client-go
// (and an optional `*http.Response`) into a `*DaytonaError`.
func ConvertAPIError(err error, httpResp *http.Response) error {
	if err == nil {
		return nil
	}
	return fromGenericError(err, httpResp,
		func(e error) ([]byte, bool) {
			var g *apiclient.GenericOpenAPIError
			if stderrors.As(e, &g) {
				return g.Body(), true
			}
			return nil, false
		})
}

// ConvertToolboxError converts an error returned by the generated
// toolbox-api-client-go into a `*DaytonaError`.
func ConvertToolboxError(err error, httpResp *http.Response) error {
	if err == nil {
		return nil
	}
	return fromGenericError(err, httpResp,
		func(e error) ([]byte, bool) {
			var g *toolbox.GenericOpenAPIError
			if stderrors.As(e, &g) {
				return g.Body(), true
			}
			return nil, false
		})
}

func fromGenericError(err error, httpResp *http.Response, extractBody func(error) ([]byte, bool)) error {
	var statusCode int
	var headers http.Header
	if httpResp != nil {
		statusCode = httpResp.StatusCode
		headers = httpResp.Header
	}

	body, hasBody := extractBody(err)
	if hasBody && len(body) > 0 {
		return NewDaytonaErrorFromBody(body, statusCode, headers)
	}

	return &DaytonaError{Message: err.Error(), StatusCode: statusCode, Headers: headers}
}
