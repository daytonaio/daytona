// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) getVolumesMountPathBinds(ctx context.Context, volumes []dto.VolumeDTO) ([]string, error) {
	volumeMountPathBinds := make([]string, 0)

	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("daytona-volume-%s", vol.VolumeId)
		runnerVolumeMountPath := d.getRunnerVolumeMountPath(volumeIdPrefixed)

		// Get or create mutex for this volume
		d.volumeMutexesMutex.Lock()
		volumeMutex, exists := d.volumeMutexes[volumeIdPrefixed]
		if !exists {
			volumeMutex = &sync.Mutex{}
			d.volumeMutexes[volumeIdPrefixed] = volumeMutex
		}
		d.volumeMutexesMutex.Unlock()

		// Lock this specific volume's mutex
		volumeMutex.Lock()
		defer volumeMutex.Unlock()

		if d.isDirectoryMounted(runnerVolumeMountPath) {
			log.Infof("volume %s is already mounted to %s", volumeIdPrefixed, runnerVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
			continue
		}

		err := os.MkdirAll(runnerVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", runnerVolumeMountPath, err)
		}

		log.Infof("mounting S3 volume %s to %s", volumeIdPrefixed, runnerVolumeMountPath)

		cmd := d.getMountCmd(ctx, volumeIdPrefixed, runnerVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeIdPrefixed, runnerVolumeMountPath, err)
		}

		log.Infof("mounted S3 volume %s to %s", volumeIdPrefixed, runnerVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) getRunnerVolumeMountPath(volumeId string) string {
	volumePath := filepath.Join("/mnt", volumeId)
	if config.GetEnvironment() == "development" {
		volumePath = filepath.Join("/tmp", volumeId)
	}

	return volumePath
}

func (d *DockerClient) isDirectoryMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	_, err := cmd.Output()

	return err == nil
}

func (d *DockerClient) getMountCmd(ctx context.Context, volume, path string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "mount-s3", "--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777", volume, path)

	if d.awsEndpointUrl != "" {
		cmd.Env = append(cmd.Env, "AWS_ENDPOINT_URL="+d.awsEndpointUrl)
	}

	if d.awsAccessKeyId != "" {
		cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY_ID="+d.awsAccessKeyId)
	}

	if d.awsSecretAccessKey != "" {
		cmd.Env = append(cmd.Env, "AWS_SECRET_ACCESS_KEY="+d.awsSecretAccessKey)
	}

	if d.awsRegion != "" {
		cmd.Env = append(cmd.Env, "AWS_REGION="+d.awsRegion)
	}

	cmd.Stderr = io.Writer(&util.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&util.InfoLogWriter{})

	return cmd
}
