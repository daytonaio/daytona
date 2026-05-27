// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ProxyErrorCode is the proxy-emitted machine-readable code stamped onto the
// wire envelope. Sites that don't warrant a code use common-go's generic
// typed errors directly.
type ProxyErrorCode string

const (
	CodeSandboxNotStarted ProxyErrorCode = "SANDBOX_NOT_STARTED"
	CodeSandboxNotFound   ProxyErrorCode = "SANDBOX_NOT_FOUND"
	CodeRunnerUnreachable ProxyErrorCode = "RUNNER_UNREACHABLE"
)

type SandboxNotStartedError struct{ Message string }

func NewSandboxNotStartedError(message string) *SandboxNotStartedError {
	return &SandboxNotStartedError{Message: message}
}
func (e *SandboxNotStartedError) Error() string       { return e.Message }
func (e *SandboxNotStartedError) HTTPStatusCode() int { return http.StatusBadRequest }
func (e *SandboxNotStartedError) ErrorCode() string   { return string(CodeSandboxNotStarted) }

type SandboxNotFoundError struct{ Message string }

func NewSandboxNotFoundError(message string) *SandboxNotFoundError {
	return &SandboxNotFoundError{Message: message}
}
func (e *SandboxNotFoundError) Error() string       { return e.Message }
func (e *SandboxNotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e *SandboxNotFoundError) ErrorCode() string   { return string(CodeSandboxNotFound) }

type RunnerUnreachableError struct{ Message string }

func NewRunnerUnreachableError(message string) *RunnerUnreachableError {
	return &RunnerUnreachableError{Message: message}
}
func (e *RunnerUnreachableError) Error() string       { return e.Message }
func (e *RunnerUnreachableError) HTTPStatusCode() int { return http.StatusBadGateway }
func (e *RunnerUnreachableError) ErrorCode() string   { return string(CodeRunnerUnreachable) }

var (
	_ common_errors.HTTPError = (*SandboxNotStartedError)(nil)
	_ common_errors.HTTPError = (*SandboxNotFoundError)(nil)
	_ common_errors.HTTPError = (*RunnerUnreachableError)(nil)
)

// classifyUpstreamError maps an api-client-go error into a proxy-flavored
// envelope. Accepts errors that may be wrapped with fmt.Errorf("...: %w", err);
// the wrapper's message becomes the wire message, while the unwrapped
// *common_errors.CustomError drives status/code classification.
func classifyUpstreamError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	var customErr *common_errors.CustomError
	if !errors.As(err, &customErr) {
		return NewRunnerUnreachableError(msg)
	}

	switch customErr.Code {
	case "SANDBOX_RUNNER_NOT_FOUND":
		return NewSandboxNotStartedError("sandbox is not running — start the sandbox before accessing it")
	case "SANDBOX_STATE_ERROR", "NO_AVAILABLE_RUNNERS":
		return NewSandboxNotStartedError(msg)
	}

	switch customErr.StatusCode {
	case http.StatusNotFound:
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "snapshot") {
			return common_errors.NewNotFoundError(errors.New(msg))
		}
		return NewSandboxNotFoundError(msg)
	case http.StatusBadRequest:
		// Back-compat: drop once all upstreams emit a typed code.
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "not started") || strings.Contains(lower, "not running") || strings.Contains(lower, "stopped") {
			return NewSandboxNotStartedError("sandbox is not running — start the sandbox before accessing it")
		}
	case http.StatusUnauthorized:
		return common_errors.NewUnauthorizedError(errors.New(msg))
	case http.StatusForbidden:
		return common_errors.NewForbiddenError(errors.New(msg))
	}

	return common_errors.NewCustomError(customErr.StatusCode, msg, "")
}

// runnerUnreachableErrorHandler fires when httputil.ReverseProxy can't dial
// the runner. Replaces httputil's default bare-502 with a typed envelope.
func runnerUnreachableErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(req.Context().Err(), context.Canceled) {
		return
	}

	log.WithFields(log.Fields{
		"path":   req.URL.Path,
		"method": req.Method,
		"error":  err,
	}).Warn("proxy upstream unreachable")

	body := common_errors.ErrorResponse{
		StatusCode: http.StatusBadGateway,
		Message:    "runner is unreachable: " + err.Error(),
		Source:     "DAYTONA_PROXY",
		Code:       string(CodeRunnerUnreachable),
		Timestamp:  time.Now(),
		Path:       req.URL.Path,
		Method:     req.Method,
	}

	payload, mErr := json.Marshal(body)
	if mErr != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusBadGateway)
	_, _ = rw.Write(payload)
}

// translateUpstreamErrorResponse rewrites runner error envelopes so the SDK
// only sees source=DAYTONA_PROXY on the preview path. Non-JSON, non-Daytona,
// and non-runner bodies pass through untouched.
func translateUpstreamErrorResponse(res *http.Response) error {
	if res == nil || res.StatusCode < 400 {
		return nil
	}
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return nil
	}

	raw, err := io.ReadAll(res.Body)
	closeErr := res.Body.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}

	var upstream common_errors.ErrorResponse
	if jsonErr := json.Unmarshal(raw, &upstream); jsonErr != nil || upstream.Source == "" {
		restoreBody(res, raw)
		return nil
	}
	if upstream.Source != "DAYTONA_RUNNER" {
		restoreBody(res, raw)
		return nil
	}

	translated := translateRunnerEnvelope(upstream, res.Request)

	payload, marshalErr := json.Marshal(translated)
	if marshalErr != nil {
		restoreBody(res, raw)
		return nil
	}

	res.StatusCode = translated.StatusCode
	res.Status = strconv.Itoa(translated.StatusCode) + " " + http.StatusText(translated.StatusCode)
	restoreBody(res, payload)
	res.Header.Set("Content-Type", "application/json")
	res.Header.Set("Content-Length", strconv.Itoa(len(payload)))
	res.Header.Del("Retry-After")
	return nil
}

func restoreBody(res *http.Response, body []byte) {
	res.Body = io.NopCloser(bytes.NewReader(body))
	res.ContentLength = int64(len(body))
	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
}

func translateRunnerEnvelope(upstream common_errors.ErrorResponse, req *http.Request) common_errors.ErrorResponse {
	const (
		runnerCodeSandboxDaemonUnreachable = "SANDBOX_DAEMON_UNREACHABLE"
		runnerCodeDockerDaemonUnreachable  = "DOCKER_DAEMON_UNREACHABLE"
	)

	statusCode := upstream.StatusCode
	code := ""
	message := upstream.Message

	switch upstream.Code {
	case runnerCodeSandboxDaemonUnreachable:
		// Forward the runner's message verbatim; only swap status + code.
		statusCode = http.StatusBadRequest
		code = string(CodeSandboxNotStarted)
	case runnerCodeDockerDaemonUnreachable:
		statusCode = http.StatusInternalServerError
		code = string(CodeRunnerUnreachable)
	}

	path := upstream.Path
	method := upstream.Method
	if req != nil {
		if path == "" && req.URL != nil {
			path = req.URL.Path
		}
		if method == "" {
			method = req.Method
		}
	}

	return common_errors.ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Source:     "DAYTONA_PROXY",
		Code:       code,
		Timestamp:  time.Now(),
		Path:       path,
		Method:     method,
	}
}
