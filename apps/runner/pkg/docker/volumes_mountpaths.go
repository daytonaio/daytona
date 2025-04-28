// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"

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

		_, err = os.Stat(nodeVolumeMountPath)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(nodeVolumeMountPath, 0755)
				if err != nil {
					return nil, fmt.Errorf("failed to create mount directory %s: %s", nodeVolumeMountPath, err)
				}

				log.Infof("created mount directory %s", nodeVolumeMountPath)
			} else {
				return nil, fmt.Errorf("failed to check mount directory %s: %s", nodeVolumeMountPath, err)
			}
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
