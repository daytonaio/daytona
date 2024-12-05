// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) ListTargets(ctx context.Context, filter *stores.TargetFilter, params services.TargetRetrievalParams) ([]services.TargetDTO, error) {
	targets, err := s.targetStore.List(filter)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []services.TargetDTO{}

	for i, t := range targets {
		state := t.GetState()

		if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
			continue
		}

		var updatedWorkspaces []models.Workspace
		for _, w := range t.Workspaces {
			wsState := w.GetState()
			if wsState.Name != models.ResourceStateNameDeleted {
				updatedWorkspaces = append(updatedWorkspaces, w)
			}
		}
		t.Workspaces = updatedWorkspaces

		response = append(response, services.TargetDTO{
			Target: *t,
			State:  state,
		})

		if !params.Verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.TargetInfoResult, 1)

			go func() {
				targetInfo, err := s.provisioner.GetTargetInfo(ctx, t)
				resultCh <- provisioner.TargetInfoResult{Info: targetInfo, Err: err}
			}()

			select {
			case res := <-resultCh:
				if res.Err != nil {
					log.Error(fmt.Errorf("failed to get target info for %s: %v", t.Name, res.Err))
					return
				}

				response[i].Info = res.Info
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					log.Warn(fmt.Sprintf("timeout getting target info for %s", t.Name))
				} else {
					log.Warn(fmt.Sprintf("cancelled getting target info for %s", t.Name))
				}
			}
		}(i)
	}

	wg.Wait()
	return response, nil
}
