// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	log "github.com/sirupsen/logrus"
)

func (tj *TargetJob) delete(ctx context.Context, j *models.Job, force bool) error {
	t, err := tj.findTarget(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	targetLogger := tj.loggerFactory.CreateTargetLogger(t.Id, t.Name, logs.LogSourceServer)

	targetLogger.Write([]byte(fmt.Sprintf("Destroying target %s", t.Name)))

	err = tj.provisioner.DestroyTarget(t)
	if err != nil {
		if !force {
			return err
		}
		log.Error(err)
	}

	targetLogger.Write([]byte(fmt.Sprintf("Target %s destroyed", t.Name)))

	err = targetLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the target logger cannot be cleaned up
		log.Error(err)
	}

	return nil
}
