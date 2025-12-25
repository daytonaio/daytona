// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) Destroy(ctx context.Context, containerId string) error {
	log.Infoln("Destroy")
	return nil
}

func (l *LibVirt) RemoveDestroyed(ctx context.Context, containerId string) error {
	log.Infoln("RemoveDestroyed")
	return nil
}
