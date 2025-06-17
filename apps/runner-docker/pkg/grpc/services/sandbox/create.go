// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/constants"
	"github.com/daytonaio/runner-docker/internal/metrics"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Sandbox endpoints
func (s *SandboxService) CreateSandbox(ctx context.Context, req *pb.CreateSandboxRequest) (*pb.CreateSandboxResponse, error) {
	startTime := time.Now()
	defer func() {
		obs, err := metrics.ContainerOperationDuration.GetMetricWithLabelValues("create")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := s.getSandboxState(ctx, req.GetId())
	if err != nil && state == pb.SandboxState_SANDBOX_STATE_ERROR {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	if state == pb.SandboxState_SANDBOX_STATE_STARTED || state == pb.SandboxState_SANDBOX_STATE_PULLING_SNAPSHOT || state == pb.SandboxState_SANDBOX_STATE_STARTING {
		metrics.SuccessCounterInc(metrics.CreateSandboxOperation)

		return &pb.CreateSandboxResponse{
			SandboxId: req.GetId(),
		}, nil
	}

	if state == pb.SandboxState_SANDBOX_STATE_STOPPED || state == pb.SandboxState_SANDBOX_STATE_CREATING {
		_, err = s.StartSandbox(ctx, &pb.StartSandboxRequest{SandboxId: req.GetId()})
		if err != nil {
			metrics.FailureCounterInc(metrics.CreateSandboxOperation)
			return nil, err
		}

		metrics.SuccessCounterInc(metrics.CreateSandboxOperation)

		return &pb.CreateSandboxResponse{
			SandboxId: req.GetId(),
		}, nil
	}

	s.cache.SetSandboxState(ctx, req.GetId(), pb.SandboxState_SANDBOX_STATE_CREATING)

	ctx = context.WithValue(ctx, constants.ID_KEY, req.GetId())

	_, err = s.snapshotService.PullSnapshot(ctx, &pb.PullSnapshotRequest{
		Snapshot: req.GetSnapshot(),
		Registry: req.GetRegistry(),
	})
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.GetId(), pb.SandboxState_SANDBOX_STATE_CREATING)

	err = s.validateImageArchitecture(ctx, req.GetSnapshot())
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	volumeMountPathBinds := make([]string, 0)
	if req.Volumes != nil {
		volumeMountPathBinds, err = s.getVolumesMountPathBinds(ctx, req.GetVolumes())
		if err != nil {
			metrics.FailureCounterInc(metrics.CreateSandboxOperation)
			return nil, err
		}
	}

	containerConfig, hostConfig, networkingConfig, err := s.getContainerConfigs(ctx, req, volumeMountPathBinds)
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	c, err := s.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, req.GetId())
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, common.MapDockerError(err)
	}

	_, err = s.StartSandbox(ctx, &pb.StartSandboxRequest{SandboxId: req.GetId()})
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	// wait for the daemon to start listening on port 2280
	container, err := s.dockerClient.ContainerInspect(ctx, c.ID)
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, common.MapDockerError(err)
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, errors.New("container has no IP address, it might not be running")
	}

	err = s.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
	if err != nil {
		metrics.FailureCounterInc(metrics.CreateSandboxOperation)
		return nil, err
	}

	metrics.SuccessCounterInc(metrics.CreateSandboxOperation)

	return &pb.CreateSandboxResponse{
		SandboxId: req.Id,
	}, nil
}

func (s *SandboxService) validateImageArchitecture(ctx context.Context, image string) error {
	inspect, err := s.dockerClient.ImageInspect(ctx, image)
	if err != nil {
		return common.MapDockerError(err)
	}

	arch := strings.ToLower(inspect.Architecture)
	validArchs := []string{"amd64", "x86_64"}

	for _, validArch := range validArchs {
		if arch == validArch {
			return nil
		}
	}

	return status.Error(codes.AlreadyExists, fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture).Error())
}

func (s *SandboxService) getContainerConfigs(ctx context.Context, req *pb.CreateSandboxRequest, volumeMountPathBinds []string) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	containerConfig := s.getContainerCreateConfig(req)

	hostConfig, err := s.getContainerHostConfig(ctx, req, volumeMountPathBinds)
	if err != nil {
		return nil, nil, nil, err
	}

	networkingConfig := s.getContainerNetworkingConfig(ctx)

	return containerConfig, hostConfig, networkingConfig, nil
}

func (s *SandboxService) getContainerCreateConfig(req *pb.CreateSandboxRequest) *container.Config {
	envVars := []string{
		"DAYTONA_WS_ID=" + req.GetId(),
		"DAYTONA_WS_IMAGE=" + req.GetSnapshot(),
		"DAYTONA_WS_USER=" + req.GetOsUser(),
	}

	for key, value := range req.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: req.GetId(),
		Image:    req.GetSnapshot(),
		// User:         sandboxDto.OsUser,
		Env:          envVars,
		Entrypoint:   req.GetEntrypoint(),
		AttachStdout: true,
		AttachStderr: true,
	}
}

func (s *SandboxService) getContainerHostConfig(ctx context.Context, req *pb.CreateSandboxRequest, volumeMountPathBinds []string) (*container.HostConfig, error) {
	var binds []string

	binds = append(binds, fmt.Sprintf("%s:/usr/local/bin/daytona:ro", s.daemonPath))

	if len(volumeMountPathBinds) > 0 {
		binds = append(binds, volumeMountPathBinds...)
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
		Resources: container.Resources{
			CPUPeriod:  100000,
			CPUQuota:   req.CpuQuota * 100000,
			Memory:     req.MemoryQuota * 1024 * 1024 * 1024,
			MemorySwap: req.MemoryQuota * 1024 * 1024 * 1024,
		},
		Binds: binds,
	}

	if s.containerRuntime != "" {
		hostConfig.Runtime = s.containerRuntime
	}

	filesystem, err := s.getFilesystem(ctx)
	if err != nil {
		return nil, err
	}

	if filesystem == "xfs" {
		hostConfig.StorageOpt = map[string]string{
			"size": fmt.Sprintf("%dG", req.StorageQuota),
		}
	}

	return hostConfig, nil
}

func (s *SandboxService) getContainerNetworkingConfig(_ context.Context) *network.NetworkingConfig {
	if s.containerNetwork != "" {
		return &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				s.containerNetwork: {},
			},
		}
	}
	return nil
}

func (s *SandboxService) getFilesystem(ctx context.Context) (string, error) {
	info, err := s.dockerClient.Info(ctx)
	if err != nil {
		return "", common.MapDockerError(err)
	}

	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			return driver[1], nil
		}
	}

	return "", errors.New("filesystem not found")
}

func (s *SandboxService) getVolumesMountPathBinds(ctx context.Context, volumes []*pb.Volume) ([]string, error) {
	volumeMountPathBinds := make([]string, 0)

	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("daytona-volume-%s", vol.VolumeId)
		nodeVolumeMountPath := s.getNodeVolumeMountPath(volumeIdPrefixed)

		// Get or create mutex for this volume
		s.volumeMutexesMutex.Lock()
		volumeMutex, exists := s.volumeMutexes[volumeIdPrefixed]
		if !exists {
			volumeMutex = &sync.Mutex{}
			s.volumeMutexes[volumeIdPrefixed] = volumeMutex
		}
		s.volumeMutexesMutex.Unlock()

		// Lock this specific volume's mutex
		volumeMutex.Lock()
		defer volumeMutex.Unlock()

		if s.isDirectoryMounted(nodeVolumeMountPath) {
			s.log.Info("volume is already mounted", "volumeIdPrefixed", volumeIdPrefixed, "nodeVolumeMountPath", nodeVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
			continue
		}

		err := os.MkdirAll(nodeVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", nodeVolumeMountPath, err)
		}

		s.log.Info("mounting S3 volume", "volumeIdPrefixed", volumeIdPrefixed, "nodeVolumeMountPath", nodeVolumeMountPath)

		cmd := s.getMountCmd(ctx, volumeIdPrefixed, nodeVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeIdPrefixed, nodeVolumeMountPath, err)
		}

		s.log.Info("mounted S3 volume", "volumeIdPrefixed", volumeIdPrefixed, "nodeVolumeMountPath", nodeVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (s *SandboxService) getNodeVolumeMountPath(volumeId string) string {
	volumePath := filepath.Join("/mnt", volumeId)
	if s.nodeEnv == "development" {
		volumePath = filepath.Join("/tmp", volumeId)
	}

	return volumePath
}

func (s *SandboxService) isDirectoryMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	_, err := cmd.Output()

	return err == nil
}

func (s *SandboxService) getMountCmd(ctx context.Context, volume, path string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "mount-s3", "--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777", volume, path)

	if s.awsEndpointUrl != "" {
		cmd.Env = append(cmd.Env, "AWS_ENDPOINT_URL="+s.awsEndpointUrl)
	}

	if s.awsAccessKeyId != "" {
		cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY_ID="+s.awsAccessKeyId)
	}

	if s.awsSecretAccessKey != "" {
		cmd.Env = append(cmd.Env, "AWS_SECRET_ACCESS_KEY="+s.awsSecretAccessKey)
	}

	if s.awsRegion != "" {
		cmd.Env = append(cmd.Env, "AWS_REGION="+s.awsRegion)
	}

	cmd.Stderr = io.Writer(&util.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&util.InfoLogWriter{})

	return cmd
}
