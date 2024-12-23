// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/views"
)

func (tj *TargetJob) create(ctx context.Context, j *models.Job) error {
	tg, err := tj.findTarget(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	targetLogger, err := tj.loggerFactory.CreateLogger(tg.Id, tg.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer targetLogger.Close()

	targetLogger.Write([]byte(fmt.Sprintf("Creating target %s (%s)\n", tg.Name, tg.Id)))

	p, err := tj.providerManager.GetProvider(tg.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*p).CreateTarget(&provider.TargetRequest{
		Target: tg,
	})
	if err != nil {
		return err
	}

	err = tj.handleSuccessfulCreation(ctx, tg.Id)
	if err != nil {
		return err
	}

	targetLogger.Write([]byte(views.GetPrettyLogLine("Target creation complete")))

	if tg.TargetConfig.ProviderInfo.AgentlessTarget {
		return nil
	}

	return tj.start(ctx, j)
}
