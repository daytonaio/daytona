// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) TagImage(ctx context.Context, sourceImage string, targetImage string) error {
	log.Infoln("TagImage")
	return nil
}
