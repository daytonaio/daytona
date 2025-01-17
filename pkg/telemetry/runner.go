// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
)

type RunnerEventName string

const (
	RunnerEventLifecycleRegistered          RunnerEventName = "runner_lifecycle_registered"
	RunnerEventLifecycleRegistrationFailed  RunnerEventName = "runner_lifecycle_registration_failed"
	RunnerEventLifecycleDeleted             RunnerEventName = "runner_lifecycle_deleted"
	RunnerEventLifecycleDeletionFailed      RunnerEventName = "runner_lifecycle_deletion_failed"
	RunnerEventProviderInstalled            RunnerEventName = "runner_provider_installed"
	RunnerEventProviderInstallationFailed   RunnerEventName = "runner_provider_installation_failed"
	RunnerEventProviderUninstalled          RunnerEventName = "runner_provider_uninstalled"
	RunnerEventProviderUninstallationFailed RunnerEventName = "runer_provider_uninstallation_failed"
	RunnerEventProviderUpdated              RunnerEventName = "runner_provider_updated"
	RunnerEventProviderUpdateFailed         RunnerEventName = "runner_provider_update_failed"
)

type runnerEvent struct {
	AbstractEvent
	runner *models.Runner
}

func NewRunnerEvent(name RunnerEventName, r *models.Runner, err error, extras map[string]interface{}) Event {
	return runnerEvent{
		runner: r,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e runnerEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.runner == nil {
		return props
	}

	props["runner_id"] = e.runner.Id
	props["is_local_runner"] = e.runner.Id == common.LOCAL_RUNNER_ID

	return props
}
