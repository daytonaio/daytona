// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	log.Infoln("PushImage")
	return nil
}
