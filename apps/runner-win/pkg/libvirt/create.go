// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, string, error) {
	log.Infoln("Create")
	return "", "", nil
}
