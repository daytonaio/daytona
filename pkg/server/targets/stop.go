// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) StopTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return err
	}

	for _, workspace := range target.Workspaces {
		//	todo: go routines
		err := s.provisioner.StopWorkspace(workspace, targetConfig)
		if err != nil {
			return err
		}
		if workspace.State != nil {
			workspace.State.Uptime = 0
			workspace.State.UpdatedAt = time.Now().Format(time.RFC1123)
		}
	}

	err = s.provisioner.StopTarget(target, targetConfig)
	if err == nil {
		err = s.targetStore.Save(target)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, targetConfig)
	event := telemetry.ServerEventTargetStopped
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetStopError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *TargetService) StopWorkspace(ctx context.Context, targetId, workspaceName string) error {
	w, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	workspace, err := w.GetWorkspace(workspaceName)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return err
	}

	err = s.provisioner.StopWorkspace(workspace, targetConfig)
	if err != nil {
		return err
	}

	if workspace.State != nil {
		workspace.State.Uptime = 0
		workspace.State.UpdatedAt = time.Now().Format(time.RFC1123)
	}

	return s.targetStore.Save(w)
}
