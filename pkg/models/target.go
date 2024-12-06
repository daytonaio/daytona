// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"github.com/daytonaio/daytona/internal/util"
)

const AGENT_UNRESPONSIVE_THRESHOLD = 30 * time.Second

type Target struct {
	Id             string            `json:"id" validate:"required" gorm:"primaryKey"`
	Name           string            `json:"name" validate:"required" gorm:"not null"`
	TargetConfigId string            `json:"targetConfigId" validate:"required" gorm:"not null"`
	TargetConfig   TargetConfig      `json:"targetConfig" validate:"required" gorm:"foreignKey:TargetConfigId"`
	ApiKey         string            `json:"-" validate:"required" gorm:"not null"`
	EnvVars        map[string]string `json:"envVars" validate:"required" gorm:"serializer:json;not null"`
	IsDefault      bool              `json:"default" validate:"required" gorm:"not null"`
	Workspaces     []Workspace       `json:"workspaces" validate:"required" gorm:"foreignKey:TargetId;references:Id"`
	Metadata       *TargetMetadata   `json:"metadata" validate:"optional" gorm:"foreignKey:TargetId;references:Id"`
	LastJob        *Job              `json:"lastJob" validate:"optional" gorm:"foreignKey:ResourceId;references:Id"`
} // @name Target

type TargetMetadata struct {
	TargetId  string    `json:"targetId" validate:"required" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required" gorm:"not null"`
	Uptime    uint64    `json:"uptime" validate:"required" gorm:"not null"`
} // @name TargetMetadata

func (t *Target) GetState() ResourceState {
	state := getResourceStateFromJob(t.LastJob)

	// Some providers do not utilize agents in target mode
	if state.Name != ResourceStateNameDeleted && t.TargetConfig.ProviderInfo.AgentlessTarget {
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
	Name            string  `json:"name" validate:"required"`
	Version         string  `json:"version" validate:"required"`
	AgentlessTarget bool    `json:"agentlessTarget" validate:"optional"`
	Label           *string `json:"label" validate:"optional"`
} // @name TargetProviderInfo
