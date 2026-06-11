// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// ProxyRequest handles proxying requests to a sandbox's container
//
//	@Tags			toolbox
//	@Summary		Proxy requests to the sandbox toolbox
//	@Description	Forwards the request to the specified sandbox's container
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Param			path		path		string	true	"Path to forward"
//	@Success		200			{object}	any		"Proxied response"
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		401			{object}	string	"Unauthorized"
//	@Failure		404			{object}	string	"Sandbox container not found"
//	@Failure		409			{object}	string	"Sandbox container conflict"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [get]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [post]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [delete]
func ProxyRequest(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Header.Get("Upgrade") != "websocket" && regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(ctx.Param("path")) {
			if ctx.Query("follow") == "true" {
				ProxyCommandLogsStream(ctx, logger)
				return
			}
		}

		sandboxId := ctx.Param("sandboxId")
		daemonPath := ctx.Param("path")
		if !strings.HasPrefix(daemonPath, "/") {
			daemonPath = "/" + daemonPath
		}

		proxy.NewProxyRequestHandler(getProxyTarget, func(resp *http.Response) error {
			// Classification is best-effort and must never fail the proxied
			// request: always return nil.
			sniffFdExhaustion(resp, sandboxId, daemonPath)
			return nil
		})(ctx)
	}
}

const (
	fdSniffErrorBodyCap = 8 * 1024
	fdSniffExecBodyCap  = 64 * 1024
)

// sniffFdExhaustion inspects small, fully-buffered daemon response bodies for
// file-descriptor exhaustion signatures and reports matches to the
// SandboxDegradedService. Streaming, chunked, or unknown-length responses are
// never buffered.
func sniffFdExhaustion(resp *http.Response, sandboxId string, daemonPath string) {
	if resp == nil || resp.Body == nil || resp.StatusCode == http.StatusSwitchingProtocols {
		return
	}

	var bodyCap int64
	switch {
	case resp.StatusCode >= 400:
		bodyCap = fdSniffErrorBodyCap
	case resp.StatusCode == http.StatusOK && daemonPath == "/process/execute":
		bodyCap = fdSniffExecBodyCap
	default:
		return
	}

	if resp.ContentLength <= 0 || resp.ContentLength > bodyCap {
		return
	}

	// ContentLength is advisory: bound the read so a misreporting daemon
	// cannot make us buffer unbounded. Reading bodyCap+1 distinguishes
	// "fits within the cap" from "longer than reported".
	body, err := io.ReadAll(io.LimitReader(resp.Body, bodyCap+1))
	if err != nil {
		// Partial read: replay what was consumed, then the read error itself.
		// The error was swallowed by our ReadAll, and the transport stream is
		// in an undefined state after a failed read — storing the error is the
		// only way the client observably sees the same failure instead of a
		// silent truncation. Skip classification.
		resp.Body = &compositeReadCloser{
			Reader: io.MultiReader(bytes.NewReader(body), &errReader{err: err}),
			Closer: resp.Body,
		}
		return
	}
	if int64(len(body)) > bodyCap {
		// The body exceeds the cap despite the reported ContentLength —
		// restore the consumed bytes ahead of the unread remainder and skip
		// classification.
		resp.Body = &compositeReadCloser{
			Reader: io.MultiReader(bytes.NewReader(body), resp.Body),
			Closer: resp.Body,
		}
		return
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	reason, ok := common.ClassifyToolboxFdExhaustion(daemonPath, resp.StatusCode, resp.Header.Get("Content-Type"), body)
	if !ok {
		return
	}

	r, err := runner.GetInstance(nil)
	if err != nil || r.SandboxDegraded == nil {
		return
	}
	r.SandboxDegraded.ReportFdExhaustion(sandboxId, reason)
}

type compositeReadCloser struct {
	io.Reader
	io.Closer
}

// errReader yields a stored error once preceding readers in a MultiReader
// chain are drained, replaying a read failure to downstream consumers.
type errReader struct{ err error }

func (e *errReader) Read([]byte) (int, error) { return 0, e.err }

func getProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return nil, nil, err
	}

	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox ID is required")))
		return nil, nil, errors.New("sandbox ID is required")
	}

	// Resolve the container IP with retries to handle transient Docker
	// networking states where the container exists but has no IP yet.
	var containerIP string
	var containerNotFound bool
	err = utils.RetryWithExponentialBackoff(ctx.Request.Context(), "resolve container IP", 3, 100*time.Millisecond, 500*time.Millisecond, func() error {
		container, err := runner.Docker.ContainerInspect(ctx.Request.Context(), sandboxId)
		if err != nil {
			containerNotFound = true
			return &utils.NonRetryableError{Err: fmt.Errorf("sandbox container not found: %w", err)}
		}

		for _, network := range container.NetworkSettings.Networks {
			containerIP = network.IPAddress
			break
		}

		if containerIP == "" {
			return errors.New("no IP address found")
		}

		return nil
	})
	if err != nil {
		if containerNotFound {
			ctx.Error(common_errors.NewNotFoundError(err))
		} else {
			ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("%w. Is the Sandbox started?", err)))
		}
		return nil, nil, err
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)

	// Get the wildcard path preserving original percent-encoding.
	// ctx.Param() decodes the path, which causes mutations when the decoded
	// form is re-encoded by Go's url package (e.g. "(" → "%28", "%40" → "@").
	path := proxy.RawParam(ctx, "path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil, nil
}
