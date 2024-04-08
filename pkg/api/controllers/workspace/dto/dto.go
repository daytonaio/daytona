// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type CreateWorkspace struct {
	Name         string
	Target       string
	Repositories []gitprovider.GitRepository
} //	@name	CreateWorkspace
