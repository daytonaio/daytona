// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package scheduler

type IScheduler interface {
	Start()
	Stop()
	AddFunc(interval string, cmd func()) error
}
