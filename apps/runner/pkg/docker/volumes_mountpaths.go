// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		runnerVolumeMountPath := d.getRunnerVolumeMountPath(volumeIdPrefixed, vol.Subpath)

		// Create unique key for this volume+subpath combination for mutex
		volumeKey := d.getVolumeKey(volumeIdPrefixed, vol.Subpath)

		// Get or create mutex for this volume+subpath
		d.volumeMutexesMutex.Lock()
		volumeMutex, exists := d.volumeMutexes[volumeKey]
		if !exists {
			volumeMutex = &sync.Mutex{}
			d.volumeMutexes[volumeKey] = volumeMutex
		}
		d.volumeMutexesMutex.Unlock()

		// Lock this specific volume's mutex
		volumeMutex.Lock()
		defer volumeMutex.Unlock()

		if d.isDirectoryMounted(runnerVolumeMountPath) {
			log.Infof("volume %s (subpath: %s) is already mounted to %s", volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
			continue
		}

		err := os.MkdirAll(runnerVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", runnerVolumeMountPath, err)
		}

		log.Infof("mounting S3 volume %s (subpath: %s) to %s", volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath)

		cmd := d.getMountCmd(ctx, volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to mount S3 volume %s (subpath: %s) to %s: %s", volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath, err)
		}

		log.Infof("mounted S3 volume %s (subpath: %s) to %s", volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) getRunnerVolumeMountPath(volumeId, subpath string) string {
	// If subpath is provided, create a unique mount point for this volume+subpath combination
	mountDirName := volumeId
	if subpath != "" {
		// Create a short hash of the subpath to keep the path reasonable
		hash := md5.Sum([]byte(subpath))
		hashStr := hex.EncodeToString(hash[:])[:8]
		mountDirName = fmt.Sprintf("%s-%s", volumeId, hashStr)
	}

	volumePath := filepath.Join("/mnt", mountDirName)
	if config.GetEnvironment() == "development" {
		volumePath = filepath.Join("/tmp", mountDirName)
	}

	return volumePath
}

func (d *DockerClient) getVolumeKey(volumeId, subpath string) string {
	if subpath == "" {
		return volumeId
	}
	return fmt.Sprintf("%s:%s", volumeId, subpath)
}

func (d *DockerClient) isDirectoryMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	_, err := cmd.Output()

	return err == nil
}

func (d *DockerClient) getMountCmd(ctx context.Context, volume, subpath, path string) *exec.Cmd {
	args := []string{"--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777"}

	if subpath != "" {
		// Ensure subpath ends with /
		prefix := subpath
		if !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		args = append(args, "--prefix", prefix)
	}

	args = append(args, volume, path)

	cmd := exec.CommandContext(ctx, "mount-s3", args...)

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
