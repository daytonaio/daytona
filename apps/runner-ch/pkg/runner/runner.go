// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"sync"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/netrules"
)

var (
	instance     *Runner
	instanceOnce sync.Once
)

// Runner holds all the services needed by the API controllers
type Runner struct {
	CHClient        *cloudhypervisor.Client
	NetRulesManager *netrules.NetRulesManager
}

// NewRunner creates a new Runner instance and sets it as the global instance
func NewRunner(chClient *cloudhypervisor.Client, netRulesManager *netrules.NetRulesManager) *Runner {
	instanceOnce.Do(func() {
		instance = &Runner{
			CHClient:        chClient,
			NetRulesManager: netRulesManager,
		}
	})
	return instance
}

// GetInstance returns the global Runner instance
// Returns nil if NewRunner has not been called
func GetInstance() *Runner {
	return instance
}
