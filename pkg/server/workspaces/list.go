// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
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
		response = append(response, dto.WorkspaceDTO{Workspace: *w})
		if !verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			target, err := s.targetStore.Find(&provider.TargetFilter{Name: &w.Target})
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
					return
				}

				response[i].Info = res.Info
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					log.Warn(fmt.Sprintf("timeout getting workspace info for %s", w.Name))
				} else {
					log.Warn(fmt.Sprintf("cancelled getting workspace info for %s", w.Name))
				}
			}
		}(i)
	}

	wg.Wait()
	return response, nil
}
