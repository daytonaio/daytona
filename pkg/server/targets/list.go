// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) ListTargets(ctx context.Context, verbose bool) ([]dto.TargetDTO, error) {
	targets, err := s.targetStore.List()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	response := []dto.TargetDTO{}

	for i, t := range targets {
		response = append(response, dto.TargetDTO{Target: *t})
		if !verbose {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &t.TargetConfig})
			if err != nil {
				log.Error(fmt.Errorf("failed to get target config for %s", t.TargetConfig))
				return
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh := make(chan provisioner.TargetInfoResult, 1)

			go func() {
				targetInfo, err := s.provisioner.GetTargetInfo(ctx, t, targetConfig)
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
