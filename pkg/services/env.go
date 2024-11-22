// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import "github.com/daytonaio/daytona/pkg/models"

type IEnvironmentVariableService interface {
	List() ([]*models.EnvironmentVariable, error)
	Save(environmentVariable *models.EnvironmentVariable) error
	Delete(key string) error
}
