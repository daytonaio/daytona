// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuilds

import "github.com/daytonaio/daytona/pkg/server/prebuilds/dto"

type IPrebuildService interface {
	CreatePrebuild() error
	ParseEvent(payload dto.WebhookEventPayload) error
}

type PrebuildService struct {
}

type PrebuildServiceConfig struct {
}

func NewPrebuildService(config PrebuildServiceConfig) IPrebuildService {
	return &PrebuildService{}
}
