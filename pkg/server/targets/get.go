// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) GetTarget(ctx context.Context, targetId string, verbose bool) (*dto.TargetDTO, error) {
	tg, err := s.targetStore.Find(targetId)
	if err != nil {
		return nil, ErrTargetNotFound
	}

	response := dto.TargetDTO{
		Target: *tg,
	}

	if !verbose {
		return &response, nil
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &tg.TargetConfig})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resultCh := make(chan provisioner.TargetInfoResult, 1)

	go func() {
		targetInfo, err := s.provisioner.GetTargetInfo(ctx, tg, targetConfig)
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
