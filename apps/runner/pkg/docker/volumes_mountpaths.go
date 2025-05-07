// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) getVolumesMountPathBinds(ctx context.Context, volumes []dto.VolumeDTO) ([]string, error) {
	volumeMountPathBinds := make([]string, 0)
	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("daytona-volume-%s", vol.VolumeId)
		nodeVolumeMountPath := d.getNodeVolumeMountPath(volumeIdPrefixed)

		mounted, err := d.isDirectoryMounted(nodeVolumeMountPath)
		if err != nil {
			log.Errorf("failed to check if volume %s is already mounted to %s: %s", volumeIdPrefixed, nodeVolumeMountPath, err)
		}

		if mounted {
			log.Infof("volume %s is already mounted to %s", volumeIdPrefixed, nodeVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
			continue
		}

		err = os.MkdirAll(nodeVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", nodeVolumeMountPath, err)
		}

		log.Infof("mounting S3 volume %s to %s", volumeIdPrefixed, nodeVolumeMountPath)

		cmd := d.getMountCmd(ctx, volumeIdPrefixed, nodeVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeIdPrefixed, nodeVolumeMountPath, err)
		}

		log.Infof("mounted S3 volume %s to %s", volumeIdPrefixed, nodeVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", nodeVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) getNodeVolumeMountPath(volumeId string) string {
	volumePath := filepath.Join("/mnt", volumeId)
	if config.GetNodeEnv() == "development" {
		volumePath = filepath.Join("/tmp", volumeId)
	}

	return volumePath
}

func (d *DockerClient) isDirectoryMounted(path string) (bool, error) {
	cmd := exec.Command("mount")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("failed to execute mount command: %w", err)
	}

	mounts := strings.Split(out.String(), "\n")
	for _, mount := range mounts {
		mountFields := strings.Fields(mount)
		if len(mountFields) >= 3 {
			// The mount point is typically the third field
			// Check if it matches the exact path we're looking for
			if mountFields[2] == path {
				return true, nil
			}
		}
	}

	return false, nil
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
