// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) RemoveImage(ctx context.Context, imageName string, force bool) error {
	log.Infoln("RemoveImage")
	return nil
}
