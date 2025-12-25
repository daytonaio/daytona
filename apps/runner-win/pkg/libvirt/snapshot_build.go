// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) BuildSnapshot(ctx context.Context, req dto.BuildSnapshotRequestDTO) error {
	log.Infoln("BuildSnapshot")
	return nil
}
