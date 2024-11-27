// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/views"
)

func (tj *TargetJob) start(ctx context.Context, j *models.Job) error {
	tg, err := tj.findTarget(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	targetLogger := tj.loggerFactory.CreateTargetLogger(tg.Id, tg.Name, logs.LogSourceServer)
	defer targetLogger.Close()

	targetLogger.Write([]byte("Starting target\n"))

	err = tj.provisioner.StartTarget(tg)
	if err != nil {
		return err
	}

	targetLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Target %s started", tg.Name))))
	return nil
}
