// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) ContainerInspect(ctx context.Context, containerId string) (ContainerJSON, error) {
	log.Infoln("ContainerInspect")
	return ContainerJSON{}, nil
}
