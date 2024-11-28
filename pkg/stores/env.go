// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type EnvironmentVariableStore interface {
	List() ([]*models.EnvironmentVariable, error)
	Save(environmentVariable *models.EnvironmentVariable) error
	Delete(key string) error
}

var (
	ErrEnvironmentVariableNotFound = errors.New("environment variable not found")
)

func IsEnvironmentVariableNotFound(err error) bool {
	return err.Error() == ErrEnvironmentVariableNotFound.Error()
}
