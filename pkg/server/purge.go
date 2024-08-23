// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"time"

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

	fmt.Println("Deleting all workspaces...")

	errCh := make(chan error)

	err := server.Start(errCh)
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		return []error{err}
	}

	workspaces, err := s.WorkspaceService.ListWorkspaces(false)
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		if !force {
			return []error{err}
		}
	}

	if err == nil {
		for _, workspace := range workspaces {
			err := s.WorkspaceService.RemoveWorkspace(ctx, workspace.Id)
			if err != nil {
				s.trackPurgeError(ctx, force, err)
				if !force {
					return []error{err}
				} else {
					fmt.Printf("Failed to delete %s: %v\n", workspace.Name, err)
				}
			} else {
				fmt.Printf("Workspace %s deleted\n", workspace.Name)
			}
		}
	} else {
		fmt.Printf("Failed to list workspaces: %v\n", err)
	}

	if s.LocalContainerRegistry != nil {
		fmt.Println("Purging local container registry...")
		err := s.LocalContainerRegistry.Purge()
		if err != nil {
			s.trackPurgeError(ctx, force, err)
			if !force {
				return []error{err}
			} else {
				fmt.Printf("Failed to purge local container registry: %v\n", err)
			}
		}
	}

	fmt.Println("Purging Tailscale server...")
	err = s.TailscaleServer.Purge()
	if err != nil {
		s.trackPurgeError(ctx, force, err)
		if !force {
			return []error{err}
		} else {
			fmt.Printf("Failed to purge Tailscale server: %v\n", err)
		}
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
	errs := s.BuildService.MarkForDeletion(nil)
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
