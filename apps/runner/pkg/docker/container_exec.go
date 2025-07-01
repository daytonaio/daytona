// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// todo send stdout for writer os.STD_OUT
func (d *DockerClient) execSync(ctx context.Context, containerId string, execOptions container.ExecOptions, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	execOptions.Env = append([]string{"DEBIAN_FRONTEND=noninteractive"}, execOptions.Env...)

	response, err := d.apiClient.ContainerExecCreate(ctx, containerId, execOptions)
	if err != nil {
		return nil, err
	}

	result, err := d.inspectExecResp(ctx, response.ID, execStartOptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DockerClient) inspectExecResp(ctx context.Context, execID string, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	resp, err := d.apiClient.ContainerExecAttach(ctx, execID, execStartOptions)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// read the output
	outputDone := make(chan error)

	outBuf := bytes.Buffer{}
	errBuf := bytes.Buffer{}

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		outMw := io.Writer(&outBuf)
		errMw := io.Writer(&errBuf)

		if d.logWriter != nil {
			outMw = io.MultiWriter(&outBuf, d.logWriter)
			errMw = io.MultiWriter(&errBuf, d.logWriter)
		}

		_, err = stdcopy.StdCopy(outMw, errMw, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return nil, err
		}
		break

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	stdout, err := io.ReadAll(&outBuf)
	if err != nil {
		return nil, err
	}
	stderr, err := io.ReadAll(&errBuf)
	if err != nil {
		return nil, err
	}

	res, err := d.apiClient.ContainerExecInspect(ctx, execID)
	if err != nil {
		return nil, err
	}

	return &ExecResult{
		ExitCode: res.ExitCode,
		StdOut:   string(stdout),
		StdErr:   string(stderr),
	}, nil
}
