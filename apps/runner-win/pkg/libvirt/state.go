// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) DeduceSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	log.Infoln("DeduceSandboxState")
	return enums.SandboxStateUnknown, nil
}
