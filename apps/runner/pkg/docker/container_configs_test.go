// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"os"
	"testing"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func TestGetContainerHostConfigAddsCDIForGpuSandbox(t *testing.T) {
	ensureRunnerConfig(t)

	d := &DockerClient{
		daemonPath: "/opt/daytona-runner/daytona-daemon",
		gpuEnabled: true,
	}

	hostConfig, err := d.getContainerHostConfig(dto.CreateSandboxDTO{GpuQuota: 1}, nil)
	if err != nil {
		t.Fatalf("getContainerHostConfig returned error: %v", err)
	}

	if len(hostConfig.Devices) != 0 {
		t.Fatalf("expected no explicit device mappings, got %d", len(hostConfig.Devices))
	}

	if len(hostConfig.DeviceRequests) != 1 {
		t.Fatalf("expected one CDI device request, got %d", len(hostConfig.DeviceRequests))
	}

	request := hostConfig.DeviceRequests[0]
	if request.Driver != "cdi" {
		t.Fatalf("expected CDI driver, got %q", request.Driver)
	}
	if len(request.DeviceIDs) != 1 || request.DeviceIDs[0] != "nvidia.com/gpu=all" {
		t.Fatalf("expected nvidia.com/gpu=all device ID, got %#v", request.DeviceIDs)
	}
}

func TestGetContainerHostConfigRejectsCpuSandboxOnGpuRunner(t *testing.T) {
	ensureRunnerConfig(t)

	d := &DockerClient{
		daemonPath: "/opt/daytona-runner/daytona-daemon",
		gpuEnabled: true,
	}

	hostConfig, err := d.getContainerHostConfig(dto.CreateSandboxDTO{GpuQuota: 0}, nil)
	if err == nil {
		t.Fatalf("expected getContainerHostConfig to reject CPU-only sandbox on GPU runner")
	}

	if hostConfig != nil {
		t.Fatalf("expected no host config for rejected CPU-only sandbox, got %#v", hostConfig)
	}
}

func TestGetContainerHostConfigAllowsCpuSandboxOnNonGpuRunner(t *testing.T) {
	ensureRunnerConfig(t)

	d := &DockerClient{
		daemonPath: "/opt/daytona-runner/daytona-daemon",
		gpuEnabled: false,
	}

	hostConfig, err := d.getContainerHostConfig(dto.CreateSandboxDTO{GpuQuota: 0}, nil)
	if err != nil {
		t.Fatalf("getContainerHostConfig returned error: %v", err)
	}

	if len(hostConfig.DeviceRequests) != 0 {
		t.Fatalf("expected no CDI device requests, got %d", len(hostConfig.DeviceRequests))
	}
	if len(hostConfig.Devices) != 0 {
		t.Fatalf("expected no explicit device mappings, got %d", len(hostConfig.Devices))
	}
}

func TestGetContainerHostConfigRejectsGpuSandboxOnNonGpuRunner(t *testing.T) {
	ensureRunnerConfig(t)

	d := &DockerClient{
		daemonPath: "/opt/daytona-runner/daytona-daemon",
		gpuEnabled: false,
	}

	hostConfig, err := d.getContainerHostConfig(dto.CreateSandboxDTO{GpuQuota: 1}, nil)
	if err == nil {
		t.Fatalf("expected getContainerHostConfig to reject GPU sandbox on non-GPU runner")
	}

	if hostConfig != nil {
		t.Fatalf("expected no host config for rejected GPU sandbox, got %#v", hostConfig)
	}
}

func ensureRunnerConfig(t *testing.T) {
	t.Helper()

	setEnvIfUnset("SERVER_URL", "https://daytona.example.test")
	setEnvIfUnset("API_TOKEN", "test-token")
	setEnvIfUnset("RUNNER_DOMAIN", "127.0.0.1")

	if _, err := config.GetConfig(); err != nil {
		t.Fatalf("failed to initialize runner config: %v", err)
	}
}

func setEnvIfUnset(key, value string) {
	if _, ok := os.LookupEnv(key); !ok {
		_ = os.Setenv(key, value)
	}
}
