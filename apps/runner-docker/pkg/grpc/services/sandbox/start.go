// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SandboxService) StartSandbox(ctx context.Context, req *pb.StartSandboxRequest) (*pb.StartSandboxResponse, error) {
	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_STARTING)

	c, err := s.dockerClient.ContainerInspect(ctx, req.GetSandboxId())
	if err != nil {
		return nil, common.MapDockerError(err)
	}

	var containerIP string
	for _, network := range c.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if c.State.Running {
		err = s.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
		if err != nil {
			return nil, err
		}

		s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_STARTED)
		return nil, nil
	}

	err = s.dockerClient.ContainerStart(ctx, req.GetSandboxId(), container.StartOptions{})
	if err != nil {
		return nil, common.MapDockerError(err)
	}

	// make sure container is running
	err = s.waitForContainerRunning(ctx, req.GetSandboxId(), 10*time.Second)
	if err != nil {
		return nil, err
	}

	processesCtx := context.Background()

	go func() {
		if err := s.startDaytonaDaemon(processesCtx, req.SandboxId); err != nil {
			s.log.Error("Failed to start Daytona daemon", "error", err)
		}
	}()

	err = s.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
	if err != nil {
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_STARTED)

	return &pb.StartSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s started", req.SandboxId),
	}, nil
}

func (s *SandboxService) waitForContainerRunning(ctx context.Context, sandboxId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return status.Errorf(codes.DeadlineExceeded, "timeout waiting for container %s to start", sandboxId)
		case <-ticker.C:
			c, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
			if err != nil {
				return common.MapDockerError(err)
			}

			if c.State.Running {
				return nil
			}
		}
	}
}

func (s *SandboxService) waitForDaemonRunning(ctx context.Context, containerIP string, timeout time.Duration) error {
	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, fmt.Errorf("failed to parse target URL: %w", err).Error())
	}

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", target.Host, timeout)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		break
	}

	return nil
}

func (s *SandboxService) startDaytonaDaemon(ctx context.Context, sandboxId string) error {
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
		s.log.Error("Error starting Daytona daemon", "error", err)
		return nil
	}

	if result.ExitCode != 0 && result.StdErr != "" {
		s.log.Error("Error starting Daytona daemon", "error", string(result.StdErr))
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
func (s *SandboxService) execSync(ctx context.Context, sandboxId string, execOptions container.ExecOptions, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	execOptions.Env = append([]string{"DEBIAN_FRONTEND=noninteractive"}, execOptions.Env...)

	response, err := s.dockerClient.ContainerExecCreate(ctx, sandboxId, execOptions)
	if err != nil {
		return nil, common.MapDockerError(err)
	}

	result, err := s.inspectExecResp(ctx, response.ID, execStartOptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SandboxService) inspectExecResp(ctx context.Context, sandboxId string, execStartOptions container.ExecStartOptions) (*ExecResult, error) {
	resp, err := s.dockerClient.ContainerExecAttach(ctx, sandboxId, execStartOptions)
	if err != nil {
		return nil, common.MapDockerError(err)
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
		return nil, common.MapDockerError(err)
	}

	return &ExecResult{
		ExitCode: res.ExitCode,
		StdOut:   string(stdout),
		StdErr:   string(stderr),
	}, nil
}
