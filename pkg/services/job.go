// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IJobService interface {
	Save(job *models.Job) error
	Find(filter *stores.JobFilter) (*models.Job, error)
	List(filter *stores.JobFilter) ([]*models.Job, error)
	Delete(job *models.Job) error
}
