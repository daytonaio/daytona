// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/services"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, params services.WorkspaceRetrievalParams) ([]services.WorkspaceDTO, error) {
	workspaces, err := s.workspaceStore.List()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []services.WorkspaceDTO{}

	for i, ws := range workspaces {
		state := ws.GetState()

		if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
			continue
		}

		response = append(response, services.WorkspaceDTO{
			Workspace: *ws,
			State:     state,
		})

		if !params.Verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.WorkspaceInfoResult, 1)

			go func() {
				workspaceInfo, err := s.provisioner.GetWorkspaceInfo(ctx, ws)
				resultCh <- provisioner.WorkspaceInfoResult{Info: workspaceInfo, Err: err}
			}()

			select {
			case res := <-resultCh:
				if res.Err != nil {
					log.Error(fmt.Errorf("failed to get workspace info for %s: %v", ws.Name, res.Err))
					return
				}

				response[i].Info = res.Info
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					log.Warn(fmt.Sprintf("timeout getting workspace info for %s", ws.Name))
				} else {
					log.Warn(fmt.Sprintf("cancelled getting workspace info for %s", ws.Name))
				}
			}
		}(i)
	}

	wg.Wait()
	return response, nil
}
