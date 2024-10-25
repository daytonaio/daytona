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

	for _, project := range target.Projects {
		//	todo: go routines
		err := s.provisioner.StopProject(project, targetConfig)
		if err != nil {
			return err
		}
		if project.State != nil {
			project.State.Uptime = 0
			project.State.UpdatedAt = time.Now().Format(time.RFC1123)
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

func (s *TargetService) StopProject(ctx context.Context, targetId, projectName string) error {
	w, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return ErrProjectNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return err
	}

	err = s.provisioner.StopProject(project, targetConfig)
	if err != nil {
		return err
	}

	if project.State != nil {
		project.State.Uptime = 0
		project.State.UpdatedAt = time.Now().Format(time.RFC1123)
	}

	return s.targetStore.Save(w)
}
