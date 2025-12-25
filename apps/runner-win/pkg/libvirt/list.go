// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) ContainerList(ctx context.Context, options ContainerListOptions) ([]ContainerSummary, error) {
	log.Infoln("ContainerList")
	return []ContainerSummary{}, nil
}

func (l *LibVirt) Info(ctx context.Context) (SystemInfo, error) {
	log.Infoln("Info")
	return SystemInfo{}, nil
}
