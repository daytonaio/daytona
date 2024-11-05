// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	db_dto "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) ListTargets(ctx context.Context, filter *target.TargetFilter, verbose bool) ([]dto.TargetDTO, error) {
	targets, err := s.targetStore.List(filter)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []dto.TargetDTO{}

	for i, t := range targets {
		response = append(response, dto.TargetDTO{TargetViewDTO: *t})
		if !verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.TargetInfoResult, 1)

			go func() {
				targetInfo, err := s.provisioner.GetTargetInfo(ctx, db_dto.ViewToTarget(t))
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
