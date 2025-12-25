// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) Start(ctx context.Context, containerId string, metadata map[string]string) (string, error) {
	log.Infoln("Start")
	return "", nil
}
