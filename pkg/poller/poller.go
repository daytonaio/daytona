// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package poller

import (
	"github.com/daytonaio/daytona/pkg/scheduler"
)

type IPoller interface {
	Start() error
	Stop()
	Poll()
}

type AbstractPoller struct {
	IPoller
	interval  string
	scheduler scheduler.IScheduler
}

func NewPoller(interval string, scheduler scheduler.IScheduler) *AbstractPoller {
	return &AbstractPoller{
		interval:  interval,
		scheduler: scheduler,
	}
}

func (p *AbstractPoller) Start() error {
	err := p.scheduler.AddFunc(p.interval, func() { p.Poll() })
	if err != nil {
		return err
	}

	p.scheduler.Start()

	return nil
}

func (p *AbstractPoller) Stop() {
	p.scheduler.Stop()
}
