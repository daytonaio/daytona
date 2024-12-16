// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
	log "github.com/sirupsen/logrus"
)

func (tj *TargetJob) delete(ctx context.Context, j *models.Job, force bool) error {
	t, err := tj.findTarget(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	targetLogger, err := tj.loggerFactory.CreateTargetLogger(t.Id, t.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer targetLogger.Close()

	targetLogger.Write([]byte(fmt.Sprintf("Destroying target %s", t.Name)))

	p, err := tj.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
	if err != nil {
		if force {
			log.Error(err)
			return nil
		}
		return err
	}

	_, err = (*p).DestroyTarget(&provider.TargetRequest{
		Target: t,
	})
	if err != nil {
		if force {
			log.Error(err)
			return nil
		}
		return err
	}

	targetLogger.Write([]byte(fmt.Sprintf("Target %s destroyed", t.Name)))

	err = targetLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the target logger cannot be cleaned up
		log.Error(err)
	}

	return nil
}
