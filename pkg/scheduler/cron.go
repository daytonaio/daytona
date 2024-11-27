// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"github.com/robfig/cron/v3"
)

// CronScheduler is a wrapper around the cron library.
// It implements the IScheduler interface.
// It is used to schedule tasks at specific intervals.
// Wrapping the cron library is necessary to enable proper mocking while testing dependent code.
type CronScheduler struct {
	cron *cron.Cron
}

func NewCronScheduler() *CronScheduler {
	return &CronScheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *CronScheduler) Start() {
	s.cron.Start()
}

func (s *CronScheduler) Stop() {
	s.cron.Stop()
}

func (s *CronScheduler) AddFunc(interval string, cmd func()) error {
	_, err := s.cron.AddFunc(interval, cmd)
	return err
}
