// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/runner-docker/cmd/config"
	"github.com/daytonaio/runner-docker/internal/constants"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) CreateSandbox(ctx context.Context, req *pb.CreateSandboxRequest) (*pb.CreateSandboxResponse, error) {
	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("create")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := s.getSandboxState(ctx, req.Id)
	if err != nil && state == enums.SandboxStateError {
		return nil, err
	}

	if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingImage || state == enums.SandboxStateStarting {
		return &pb.CreateSandboxResponse{
			SandboxId: req.Id,
		}, nil
	}

	if state == enums.SandboxStateStopped || state == enums.SandboxStateCreating {
		_, err = s.StartSandbox(ctx, &pb.StartSandboxRequest{SandboxId: req.Id})
		if err != nil {
			return nil, err
		}

		return &pb.CreateSandboxResponse{
			SandboxId: req.Id,
		}, nil
	}

	s.cache.SetSandboxState(ctx, req.Id, enums.SandboxStateCreating)

	ctx = context.WithValue(ctx, constants.ID_KEY, req.Id)
	_, err = s.PullImage(ctx, &pb.PullImageRequest{
		Image:    req.Image,
		Registry: req.Registry,
	})
	if err != nil {
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.Id, enums.SandboxStateCreating)

	err = s.validateImageArchitecture(ctx, req.Image)
	if err != nil {
		return nil, err
	}

	volumeMountPathBinds := make([]string, 0)
	if req.Volumes != nil {
		volumeMountPathBinds, err = s.getVolumesMountPathBinds(ctx, req.Volumes)
		if err != nil {
			return nil, err
		}
	}

	containerConfig, hostConfig, err := s.getContainerConfigs(ctx, req, volumeMountPathBinds)
	if err != nil {
		return nil, err
	}

	c, err := s.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, req.Id)
	if err != nil {
		return nil, err
	}

	_, err = s.StartSandbox(ctx, &pb.StartSandboxRequest{SandboxId: req.Id})
	if err != nil {
		return nil, err
	}

	// wait for the daemon to start listening on port 2280
	container, err := s.dockerClient.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, common.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err))
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return nil, errors.New("container has no IP address, it might not be running")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		break
	}

	return &pb.CreateSandboxResponse{
		SandboxId: req.Id,
	}, nil
}

func (s *RunnerServer) validateImageArchitecture(ctx context.Context, image string) error {
	inspect, err := s.dockerClient.ImageInspect(ctx, image)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	arch := strings.ToLower(inspect.Architecture)
	validArchs := []string{"amd64", "x86_64"}

	for _, validArch := range validArchs {
		if arch == validArch {
			return nil
		}
	}

	return common.NewConflictError(fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture))
}

func (s *RunnerServer) getContainerConfigs(ctx context.Context, req *pb.CreateSandboxRequest, volumeMountPathBinds []string) (*container.Config, *container.HostConfig, error) {
	containerConfig := s.getContainerCreateConfig(req)

	hostConfig, err := s.getContainerHostConfig(ctx, req, volumeMountPathBinds)
	if err != nil {
		return nil, nil, err
	}

	return containerConfig, hostConfig, nil
}

func (s *RunnerServer) getContainerCreateConfig(req *pb.CreateSandboxRequest) *container.Config {
	envVars := []string{
		"DAYTONA_WS_ID=" + req.Id,
		"DAYTONA_WS_IMAGE=" + req.Image,
		"DAYTONA_WS_USER=" + req.OsUser,
	}

	for key, value := range req.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: req.Id,
		Image:    req.Image,
		// User:         sandboxDto.OsUser,
		Env:          envVars,
		Entrypoint:   req.Entrypoint,
		AttachStdout: true,
		AttachStderr: true,
	}
}

func (s *RunnerServer) getContainerHostConfig(ctx context.Context, req *pb.CreateSandboxRequest, volumeMountPathBinds []string) (*container.HostConfig, error) {
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

	containerRuntime := config.GetContainerRuntime()
	if containerRuntime != "" {
		hostConfig.Runtime = containerRuntime
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

func (s *RunnerServer) getFilesystem(ctx context.Context) (string, error) {
	info, err := s.dockerClient.Info(ctx)
	if err != nil {
		return "", err
	}

	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			return driver[1], nil
		}
	}

	return "", errors.New("filesystem not found")
}

func (s *RunnerServer) getVolumesMountPathBinds(ctx context.Context, volumes []*pb.Volume) ([]string, error) {
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
			log.Infof("volume %s is already mounted to %s", volumeIdPrefixed, nodeVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
			continue
		}

		err := os.MkdirAll(nodeVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", nodeVolumeMountPath, err)
		}

		log.Infof("mounting S3 volume %s to %s", volumeIdPrefixed, nodeVolumeMountPath)

		cmd := s.getMountCmd(ctx, volumeIdPrefixed, nodeVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeIdPrefixed, nodeVolumeMountPath, err)
		}

		log.Infof("mounted S3 volume %s to %s", volumeIdPrefixed, nodeVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (s *RunnerServer) getNodeVolumeMountPath(volumeId string) string {
	volumePath := filepath.Join("/mnt", volumeId)
	if config.GetNodeEnv() == "development" {
		volumePath = filepath.Join("/tmp", volumeId)
	}

	return volumePath
}

func (s *RunnerServer) isDirectoryMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	_, err := cmd.Output()

	return err == nil
}

func (s *RunnerServer) getMountCmd(ctx context.Context, volume, path string) *exec.Cmd {
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
