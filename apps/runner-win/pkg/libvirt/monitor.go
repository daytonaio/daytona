// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/netrules"
	log "github.com/sirupsen/logrus"
)

type LibVirtMonitor struct {
	ctx             context.Context
	cancel          context.CancelFunc
	netRulesManager *netrules.NetRulesManager
}

func NewLibVirtMonitor(netRulesManager *netrules.NetRulesManager) *LibVirtMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &LibVirtMonitor{
		ctx:             ctx,
		cancel:          cancel,
		netRulesManager: netRulesManager,
	}
}

func (lm *LibVirtMonitor) Stop() {
	log.Infoln("LibVirtMonitor.Stop")
	lm.cancel()
}

func (lm *LibVirtMonitor) Start() error {
	log.Infoln("LibVirtMonitor.Start")
	return nil
}
