// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package scheduler

import "github.com/robfig/cron/v3"

type IScheduler interface {
	Start()
	Stop()
	AddFunc(interval string, cmd func()) error
}

type AbstractScheduler struct {
	IScheduler
	cron *cron.Cron
}

func NewAbstractScheduler() *AbstractScheduler {
	return &AbstractScheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *AbstractScheduler) Start() {
	s.cron.Start()
}

func (s *AbstractScheduler) Stop() {
	s.cron.Stop()
}

func (s *AbstractScheduler) AddFunc(interval string, cmd func()) error {
	_, err := s.cron.AddFunc(interval, cmd)
	return err
}
