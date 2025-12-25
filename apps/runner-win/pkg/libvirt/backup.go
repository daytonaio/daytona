// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) CreateBackup(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	log.Infoln("CreateBackup")
	return nil
}

func (l *LibVirt) CreateBackupAsync(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	log.Infoln("CreateBackupAsync")
	return nil
}
