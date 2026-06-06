// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (c *ComputerUse) waitForProcessReadiness(process *Process) error {
	timeout := process.readinessTimeout
	if timeout <= 0 {
		timeout = defaultReadinessTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := waitForReadiness(ctx, process.readinessProbe); err != nil {
		return fmt.Errorf("process %s did not become ready within %s: %w", process.Name, timeout, err)
	}
	return nil
}

func waitForReadiness(ctx context.Context, probe func(context.Context) error) error {
	var lastErr error
	for {
		if ctx.Err() != nil {
			if lastErr != nil {
				return lastErr
			}
			return ctx.Err()
		}

		if err := probe(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
		case <-time.After(readinessPollInterval):
		}
	}
}

func xDisplayReadinessProbe(display string) func(context.Context) error {
	return func(ctx context.Context) error {
		if socketPath, ok := xDisplaySocketPath(display); ok {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "unix", socketPath)
			if err != nil {
				return fmt.Errorf("X display socket %s is not ready: %w", socketPath, err)
			}
			return conn.Close()
		}

		return runReadinessCommand(ctx, map[string]string{"DISPLAY": display}, "xdpyinfo", "-display", display)
	}
}

func xDisplaySocketPath(display string) (string, bool) {
	if !strings.HasPrefix(display, ":") {
		return "", false
	}

	number := strings.TrimPrefix(display, ":")
	if dot := strings.IndexByte(number, '.'); dot >= 0 {
		number = number[:dot]
	}
	if number == "" {
		return "", false
	}
	for _, r := range number {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return filepath.Join("/tmp/.X11-unix", "X"+number), true
}

func xfce4ReadinessProbe(display string) func(context.Context) error {
	return func(ctx context.Context) error {
		output, err := runReadinessCommandOutput(ctx, map[string]string{"DISPLAY": display}, "xprop", "-root", "_NET_SUPPORTING_WM_CHECK")
		if err != nil {
			return err
		}

		return xfce4WindowManagerReady(strings.TrimSpace(string(output)))
	}
}

func xfce4WindowManagerReady(xpropOutput string) error {
	if strings.Contains(xpropOutput, "not found") || !strings.Contains(xpropOutput, "_NET_SUPPORTING_WM_CHECK(WINDOW): window id #") {
		return fmt.Errorf("xfce4 window manager is not ready: %s", xpropOutput)
	}
	return nil
}

func tcpReadinessProbe(host, port string) func(context.Context) error {
	address := net.JoinHostPort(host, port)
	return func(ctx context.Context) error {
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return fmt.Errorf("tcp %s is not ready: %w", address, err)
		}
		return conn.Close()
	}
}

func runReadinessCommand(ctx context.Context, env map[string]string, command string, args ...string) error {
	_, err := runReadinessCommandOutput(ctx, env, command, args...)
	return err
}

func runReadinessCommandOutput(ctx context.Context, env map[string]string, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = mergedEnv(env)
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text != "" {
			return output, fmt.Errorf("%s %s failed: %w: %s", command, strings.Join(args, " "), err, text)
		}
		return output, fmt.Errorf("%s %s failed: %w", command, strings.Join(args, " "), err)
	}
	return output, nil
}

func mergedEnv(extra map[string]string) []string {
	if len(extra) == 0 {
		return nil
	}
	env := os.Environ()
	for key, value := range extra {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}
