// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *Server) Purge(ctx context.Context, force bool) []error {
	log.SetLevel(log.PanicLevel)

	telemetryEnabled := telemetry.TelemetryEnabled(ctx)
	telemetryProps := map[string]interface{}{
		"force":     force,
		"server_id": s.Id,
	}

	if telemetryEnabled {
		err := s.TelemetryService.TrackServerEvent(telemetry.ServerEventPurgeStarted, telemetry.ClientId(ctx), telemetryProps)
		if err != nil {
			log.Trace(err)
		}
	}

	fmt.Println("Deleting all targets...")

	err := server.Start()
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		return []error{err}
	}

	targets, err := s.TargetService.ListTargets(ctx, nil, services.TargetRetrievalParams{})
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		if !force {
			return []error{err}
		}
	}

	if err == nil {
		for _, target := range targets {
			err := s.TargetService.RemoveTarget(ctx, target.Id)
			if err != nil {
				s.trackPurgeError(ctx, force, err)
				if !force {
					return []error{err}
				} else {
					fmt.Printf("Failed to delete %s: %v\n", target.Name, err)
				}
			} else {
				fmt.Printf("Target %s deleted\n", target.Name)
			}
		}
	} else {
		fmt.Printf("Failed to list targets: %v\n", err)
	}

	fmt.Println("Purging providers...")
	err = s.ProviderManager.Purge()
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		if !force {
			return []error{err}
		} else {
			fmt.Printf("Failed to purge providers: %v\n", err)
		}
	}

	fmt.Println("Purging builds...")
	errs := s.BuildService.MarkForDeletion(nil, force)
	if len(errs) > 0 {
		s.trackPurgeError(ctx, force, errs[0])
		if !force {
			return errs
		} else {
			fmt.Printf("Failed to mark builds for deletion: %v\n", errs[0])
		}
	}

	err = s.BuildService.AwaitEmptyList(time.Minute)
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		if !force {
			return []error{err}
		} else {
			fmt.Printf("Failed to await empty build list: %v\n", err)
		}
	}

	if telemetryEnabled {
		err := s.TelemetryService.TrackServerEvent(telemetry.ServerEventPurgeCompleted, telemetry.ClientId(ctx), telemetryProps)
		if err != nil {
			log.Trace(err)
		}
	}

	return nil
}

func (s *Server) trackPurgeError(ctx context.Context, force bool, err error) {
	telemetryEnabled := telemetry.TelemetryEnabled(ctx)
	telemetryProps := map[string]interface{}{
		"server_id": s.Id,
		"force":     force,
		"error":     err.Error(),
	}

	if telemetryEnabled {
		err := s.TelemetryService.TrackServerEvent(telemetry.ServerEventPurgeError, telemetry.ClientId(ctx), telemetryProps)
		if err != nil {
			log.Trace(err)
		}
	}
}
