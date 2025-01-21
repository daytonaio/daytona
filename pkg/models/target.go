// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"github.com/daytonaio/daytona/internal/util"
)

const RESOURCE_UNRESPONSIVE_THRESHOLD = 30 * time.Second

type Target struct {
	Id               string            `json:"id" validate:"required" gorm:"primaryKey"`
	Name             string            `json:"name" validate:"required" gorm:"not null"`
	TargetConfigId   string            `json:"targetConfigId" validate:"required" gorm:"not null"`
	TargetConfig     TargetConfig      `json:"targetConfig" validate:"required" gorm:"foreignKey:TargetConfigId"`
	ApiKey           string            `json:"-" validate:"required" gorm:"not null"`
	EnvVars          map[string]string `json:"envVars" validate:"required" gorm:"serializer:json;not null"`
	IsDefault        bool              `json:"default" validate:"required" gorm:"not null"`
	Workspaces       []Workspace       `json:"workspaces" validate:"required"`
	Metadata         *TargetMetadata   `json:"metadata" validate:"optional" gorm:"foreignKey:Id;references:TargetId"`
	LastJob          *Job              `json:"lastJob" validate:"optional" gorm:"foreignKey:Id;references:ResourceId"`
	ProviderMetadata *string           `json:"providerMetadata,omitempty" validate:"optional"`
} // @name Target

type TargetMetadata struct {
	TargetId  string    `json:"targetId" validate:"required" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required" gorm:"not null"`
	Uptime    uint64    `json:"uptime" validate:"required" gorm:"not null"`
} // @name TargetMetadata

var allowedAgentlessTargetStates = map[ResourceStateName]bool{
	ResourceStateNamePendingCreate: true,
	ResourceStateNameCreating:      true,
	ResourceStateNameError:         true,
	ResourceStateNameDeleted:       true,
}

func (t *Target) GetState() ResourceState {
	state := getResourceStateFromJob(t.LastJob)

	// Some providers do not utilize agents in target mode
	if t.TargetConfig.ProviderInfo.AgentlessTarget {
		if allowedAgentlessTargetStates[state.Name] {
			return state
		}

		return ResourceState{
			Name:      ResourceStateNameUndefined,
			UpdatedAt: time.Now(),
		}
	}

	// If the target should be running, check if it is unresponsive
	if state.Name == ResourceStateNameStarted {
		if t.Metadata != nil && time.Since(t.Metadata.UpdatedAt) > RESOURCE_UNRESPONSIVE_THRESHOLD {
			state.Name = ResourceStateNameUnresponsive
			state.Error = util.Pointer("Target is unresponsive")
			state.UpdatedAt = t.Metadata.UpdatedAt
		}
	}

	return state
}
