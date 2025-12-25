// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) UpdateNetworkSettings(ctx context.Context, containerId string, updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO) error {
	log.Infoln("UpdateNetworkSettings")
	return nil
}
