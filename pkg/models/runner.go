// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"github.com/daytonaio/daytona/internal/util"
)

type Runner struct {
	Id       string          `json:"id" validate:"required" gorm:"primaryKey"`
	Name     string          `json:"name" validate:"required" gorm:"uniqueIndex;not null"`
	ApiKey   string          `json:"-" validate:"required" gorm:"not null"`
	Metadata *RunnerMetadata `json:"metadata" validate:"optional" gorm:"foreignKey:RunnerId;references:Id"`
} // @name Runner

func (r *Runner) GetState() ResourceState {
	var state ResourceState
	state.Name = ResourceStateNameUnresponsive
	state.UpdatedAt = time.Now()
	state.Error = util.Pointer("Runner is unresponsive")

	if r.Metadata == nil {
		return state
	}

	if r.Metadata != nil && (time.Since(r.Metadata.UpdatedAt) > RESOURCE_UNRESPONSIVE_THRESHOLD || r.Metadata.Uptime == 0) {
		state.UpdatedAt = r.Metadata.UpdatedAt
		return state
	}

	state.Name = ResourceStateNameStarted
	return state
}

type RunnerMetadata struct {
	RunnerId    string         `json:"runnerId" validate:"required" gorm:"primaryKey"`
	UpdatedAt   time.Time      `json:"updatedAt" validate:"required" gorm:"not null"`
	Uptime      uint64         `json:"uptime" validate:"required" gorm:"not null"`
	RunningJobs *uint64        `json:"runningJobs,omitempty" validate:"optional" gorm:"default:0"`
	Providers   []ProviderInfo `json:"providers" validate:"required" gorm:"serializer:json;not null"`
} // @name RunnerMetadata
