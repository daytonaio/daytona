// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) ImageExists(ctx context.Context, imageName string, includeLatest bool) (bool, error) {
	log.Infoln("ImageExists")
	return false, nil
}
