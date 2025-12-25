// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type ImageInfo struct {
	Size       int64
	Entrypoint []string
	Cmd        []string
	Hash       string
}

func (l *LibVirt) GetImageInfo(ctx context.Context, imageName string) (*ImageInfo, error) {
	log.Infoln("GetImageInfo")
	return &ImageInfo{}, nil
}
