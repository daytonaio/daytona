// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"sync"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/daytonaio/runner-android/pkg/netrules"
)

var (
	instance     *Runner
	instanceOnce sync.Once
)

// Runner holds all the services needed by the API controllers
type Runner struct {
	CVDClient       *cuttlefish.Client
	NetRulesManager *netrules.NetRulesManager
}

// NewRunner creates a new Runner instance and sets it as the global instance
func NewRunner(cvdClient *cuttlefish.Client, netRulesManager *netrules.NetRulesManager) *Runner {
	instanceOnce.Do(func() {
		instance = &Runner{
			CVDClient:       cvdClient,
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
