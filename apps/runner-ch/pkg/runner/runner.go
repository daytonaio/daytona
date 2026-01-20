// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/netrules"
)

// Runner holds all the services needed by the API controllers
type Runner struct {
	CHClient        *cloudhypervisor.Client
	NetRulesManager *netrules.NetRulesManager
}

// NewRunner creates a new Runner instance
func NewRunner(chClient *cloudhypervisor.Client, netRulesManager *netrules.NetRulesManager) *Runner {
	return &Runner{
		CHClient:        chClient,
		NetRulesManager: netRulesManager,
	}
}
