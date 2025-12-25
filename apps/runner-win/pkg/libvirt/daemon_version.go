// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) GetDaemonVersion(ctx context.Context, sandboxId string) (string, error) {
	log.Infoln("GetDaemonVersion")
	return "", nil
}
