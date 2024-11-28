// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"slices"
	"time"

	"github.com/daytonaio/daytona/internal/util"
)

const AGENT_UNRESPONSIVE_THRESHOLD = 30 * time.Second

var providersWithoutTargetMode = []string{"docker-provider"}

type Target struct {
	Id             string            `json:"id" validate:"required" gorm:"primaryKey"`
	Name           string            `json:"name" validate:"required"`
	TargetConfigId string            `json:"targetConfigId" validate:"required" gorm:"foreignKey:TargetConfigId;references:Id"`
	TargetConfig   TargetConfig      `json:"targetConfig" validate:"required" gorm:"foreignKey:TargetConfigId"`
	ApiKey         string            `json:"-"`
	EnvVars        map[string]string `json:"envVars" validate:"required" gorm:"serializer:json"`
	IsDefault      bool              `json:"default" validate:"required"`
	Workspaces     []Workspace       `gorm:"foreignKey:TargetId;references:Id"`
	Metadata       *TargetMetadata   `gorm:"foreignKey:TargetId;references:Id" validate:"optional"`
	LastJob        *Job              `gorm:"foreignKey:ResourceId;references:Id" validate:"optional"`
} // @name Target

type TargetMetadata struct {
	TargetId  string    `json:"targetId" validate:"required" gorm:"primaryKey;foreignKey:TargetId;references:Id"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
	Uptime    uint64    `json:"uptime" validate:"required"`
} // @name TargetMetadata

func (t *Target) GetState() ResourceState {
	state := getResourceStateFromJob(t.LastJob)

	// Some providers do not utilize agents in target mode
	if state.Name != ResourceStateNameDeleted && slices.Contains(providersWithoutTargetMode, t.TargetConfig.ProviderInfo.Name) {
		return ResourceState{
			Name:      ResourceStateNameUndefined,
			UpdatedAt: time.Now(),
		}
	}

	// If the target should be running, check if it is unresponsive
	if state.Name == ResourceStateNameStarted {
		if t.Metadata != nil && time.Since(t.Metadata.UpdatedAt) > AGENT_UNRESPONSIVE_THRESHOLD {
			state.Name = ResourceStateNameUnresponsive
			state.Error = util.Pointer("Target is unresponsive")
			state.UpdatedAt = t.Metadata.UpdatedAt
		}
	}

	return state
}

type TargetInfo struct {
	Name             string `json:"name" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
} // @name TargetInfo

type ProviderInfo struct {
	Name    string  `json:"name" validate:"required"`
	Version string  `json:"version" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
} // @name TargetProviderInfo
