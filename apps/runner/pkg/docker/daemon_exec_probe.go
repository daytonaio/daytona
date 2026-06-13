// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type daemonExecRequest struct {
	Command string `json:"command"`
	Timeout uint32 `json:"timeout"`
}

type daemonExecResponse struct {
	ExitCode int    `json:"exitCode"`
	Result   string `json:"result"`
}

const daemonExecProbeBodyLimit = 64 * 1024

// ProbeDaemonExec checks whether the sandbox daemon can spawn a process by
// executing the shell builtin `true` via POST /process/execute (the same path
// user execs take). It takes the caller's container inspect response (the
// probe loop already inspected this tick — re-inspecting here would double
// the docker API load). It returns healthy=true when the command exits 0,
// healthy=false with the observed error text when the daemon responded but
// the exec failed, and a non-nil error when the outcome is indeterminate
// (IP not resolvable, transport or decode failure).
func (d *DockerClient) ProbeDaemonExec(ctx context.Context, c *container.InspectResponse) (bool, string, error) {
	containerIP := GetContainerIpAddress(ctx, c)
	if containerIP == "" {
		return false, "", errors.New("sandbox IP not found? Is the sandbox started?")
	}

	body, err := json.Marshal(daemonExecRequest{Command: "true", Timeout: 5})
	if err != nil {
		return false, "", err
	}

	targetUrl := fmt.Sprintf("http://%s:2280/process/execute", containerIP)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetUrl, bytes.NewReader(body))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, daemonExecProbeBodyLimit))
	if err != nil {
		return false, "", err
	}

	if resp.StatusCode != http.StatusOK {
		// The daemon responded but could not execute the command; surface the
		// error body so fd-exhaustion signatures can be matched.
		return false, string(respBody), nil
	}

	var execResp daemonExecResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return false, "", err
	}

	if execResp.ExitCode == 0 {
		return true, "", nil
	}

	return false, execResp.Result, nil
}
