// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"

	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) StartSandbox(ctx context.Context, req *pb.StartSandboxRequest) (*pb.StartSandboxResponse, error) {
	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateStarting)

	c, err := s.dockerClient.ContainerInspect(ctx, req.SandboxId)
	if err != nil {
		return nil, err
	}

	if c.State.Running {
		s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateStarted)
		return nil, nil
	}

	err = s.dockerClient.ContainerStart(ctx, req.SandboxId, container.StartOptions{})
	if err != nil {
		return nil, err
	}

	// make sure container is running
	err = s.waitForContainerRunning(ctx, req.SandboxId, 10*time.Second)
	if err != nil {
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateStarted)

	processesCtx := context.Background()

	go func() {
		if err := s.startDaytonaDaemon(processesCtx, req.SandboxId); err != nil {
			log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
		}
	}()

	return &pb.StartSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s started", req.SandboxId),
	}, nil
}

func (s *RunnerServer) waitForContainerRunning(ctx context.Context, sandboxId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to start", sandboxId)
		case <-ticker.C:
			c, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
			if err != nil {
				return err
			}

			if c.State.Running {
				return nil
			}
		}
	}
}

func (s *RunnerServer) startDaytonaDaemon(ctx context.Context, sandboxId string) error {
	execOptions := container.ExecOptions{
		Cmd:          []string{"sh", "-c", "daytona"},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execStartOptions := container.ExecStartOptions{
		Detach: false,
	}

	result, err := s.execSync(ctx, sandboxId, execOptions, execStartOptions)
	if err != nil {
		log.Errorf("Error starting Daytona daemon: %s", err.Error())
		return nil
	}

	if result.ExitCode != 0 && result.StdErr != "" {
		log.Errorf("Error starting Daytona daemon: %s", string(result.StdErr))
		return nil
	}

	return nil
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// todo send stdout for writer os.STD_OUT
func (s *RunnerServer) execSync(ctx context.Context, sandboxId string, execOptions container.ExecOptions, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	execOptions.Env = append([]string{"DEBIAN_FRONTEND=noninteractive"}, execOptions.Env...)

	response, err := s.dockerClient.ContainerExecCreate(ctx, sandboxId, execOptions)
	if err != nil {
		return nil, err
	}

	result, err := s.inspectExecResp(ctx, response.ID, execStartOptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *RunnerServer) inspectExecResp(ctx context.Context, sandboxId string, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	resp, err := s.dockerClient.ContainerExecAttach(ctx, sandboxId, execStartOptions)
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

		if s.logWriter != nil {
			outMw = io.MultiWriter(&outBuf, s.logWriter)
			errMw = io.MultiWriter(&errBuf, s.logWriter)
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

	res, err := s.dockerClient.ContainerExecInspect(ctx, sandboxId)
	if err != nil {
		return nil, err
	}

	return &ExecResult{
		ExitCode: res.ExitCode,
		StdOut:   string(stdout),
		StdErr:   string(stderr),
	}, nil
}
