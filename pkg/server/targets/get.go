// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) GetTarget(ctx context.Context, filter *stores.TargetFilter, params services.TargetRetrievalParams) (*services.TargetDTO, error) {
	tg, err := s.targetStore.Find(ctx, filter)
	if err != nil {
		return nil, stores.ErrTargetNotFound
	}

	state := tg.GetState()

	if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
		return nil, stores.ErrTargetNotFound
	}

	var updatedWorkspaces []models.Workspace
	for _, w := range tg.Workspaces {
		wsState := w.GetState()
		if wsState.Name != models.ResourceStateNameDeleted {
			updatedWorkspaces = append(updatedWorkspaces, w)
		}
	}
	tg.Workspaces = updatedWorkspaces

	response := services.TargetDTO{
		Target: *tg,
		State:  state,
	}

	if !params.Verbose {
		return &response, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resultCh := make(chan provisioner.TargetInfoResult, 1)

	go func() {
		targetInfo, err := s.provisioner.GetTargetInfo(ctx, tg)
		resultCh <- provisioner.TargetInfoResult{Info: targetInfo, Err: err}
	}()

	select {
	case res := <-resultCh:
		if res.Err != nil {
			log.Error(fmt.Errorf("failed to get target info for %s: %v", tg.Name, res.Err))
			return nil, res.Err
		}

		response.Info = res.Info
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Warn(fmt.Sprintf("timeout getting target info for %s", tg.Name))
		} else {
			log.Warn(fmt.Sprintf("cancelled getting target info for %s", tg.Name))
		}
	}

	return &response, nil
}
