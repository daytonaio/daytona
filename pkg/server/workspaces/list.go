// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, verbose bool) ([]dto.WorkspaceDTO, error) {
	workspaces, err := s.workspaceStore.List()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []dto.WorkspaceDTO{}

	for i, w := range workspaces {
		if !verbose {
			response = append(response, dto.WorkspaceDTO{Workspace: *w})
			continue
		}

		wg.Add(1)
		go func(i int, w *workspace.Workspace) {
			defer wg.Done()

			workspaceDto := dto.WorkspaceDTO{Workspace: *w}

			target, err := s.targetStore.Find(w.Target)
			if err != nil {
				log.Error(fmt.Errorf("failed to get target for %s", w.Target))
				return
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.InfoResult, 1)

			go func() {
				workspaceInfo, err := s.provisioner.GetWorkspaceInfo(ctx, w, target)
				resultCh <- provisioner.InfoResult{Info: workspaceInfo, Err: err}
			}()

			select {
			case res := <-resultCh:
				if res.Err != nil {
					log.Error(fmt.Errorf("failed to get workspace info for %s: %v", w.Name, res.Err))
					response = append(response, workspaceDto)
					return
				}

				workspaceDto.Info = res.Info
				response = append(response, workspaceDto)
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					log.Warn(fmt.Sprintf("timeout getting workspace info for %s", w.Name))
				} else {
					log.Warn(fmt.Sprintf("cancelled getting workspace info for %s", w.Name))
				}
				response = append(response, workspaceDto)
			}
		}(i, w)
	}

	wg.Wait()
	return response, nil
}
