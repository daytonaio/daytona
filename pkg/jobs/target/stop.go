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

func (tj *TargetJob) stop(ctx context.Context, j *models.Job) error {
	t, err := tj.findTarget(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	targetLogger, err := tj.loggerFactory.CreateLogger(t.Id, t.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer targetLogger.Close()

	targetLogger.Write([]byte(fmt.Sprintf("Stopping target %s\n", t.Name)))

	p, err := tj.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*p).StopTarget(&provider.TargetRequest{
		Target: t,
	})
	if err != nil {
		return err
	}

	targetLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Target %s stopped", t.Name))))

	return nil
}
