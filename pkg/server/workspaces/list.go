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
	"github.com/daytonaio/daytona/pkg/target"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, verbose bool) ([]dto.WorkspaceDTO, error) {
	workspaces, err := s.workspaceStore.List()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []dto.WorkspaceDTO{}

	for i, ws := range workspaces {
		response = append(response, dto.WorkspaceDTO{Workspace: *ws})
		if !verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &ws.TargetId})
			if err != nil {
				log.Error(fmt.Errorf("failed to get target for %s", ws.TargetId))
				return
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.WorkspaceInfoResult, 1)

			go func() {
				targetInfo, err := s.provisioner.GetWorkspaceInfo(ctx, ws, target)
				resultCh <- provisioner.WorkspaceInfoResult{Info: targetInfo, Err: err}
			}()

			select {
			case res := <-resultCh:
				if res.Err != nil {
					log.Error(fmt.Errorf("failed to get target info for %s: %v", ws.Name, res.Err))
					return
				}

				response[i].Info = res.Info
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					log.Warn(fmt.Sprintf("timeout getting target info for %s", ws.Name))
				} else {
					log.Warn(fmt.Sprintf("cancelled getting target info for %s", ws.Name))
				}
			}
		}(i)
	}

	wg.Wait()
	return response, nil
}
