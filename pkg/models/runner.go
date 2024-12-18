// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import "time"

type Runner struct {
	Id       string          `json:"id" validate:"required" gorm:"primaryKey"`
	Name     string          `json:"name" validate:"required" gorm:"uniqueIndex;not null"`
	ApiKey   string          `json:"-" validate:"required" gorm:"not null"`
	Metadata *RunnerMetadata `json:"metadata" validate:"optional" gorm:"foreignKey:RunnerId;references:Id"`
} // @name Runner

func (r *Runner) GetState() RunnerState {
	if r.Metadata == nil {
		return RunnerStateUnresponsive
	}

	if r.Metadata != nil && time.Since(r.Metadata.UpdatedAt) > AGENT_UNRESPONSIVE_THRESHOLD {
		return RunnerStateUnresponsive
	}

	return RunnerStateRunning
}

type RunnerMetadata struct {
	RunnerId    string         `json:"runnerId" validate:"required" gorm:"primaryKey"`
	UpdatedAt   time.Time      `json:"updatedAt" validate:"required" gorm:"not null"`
	Uptime      uint64         `json:"uptime" validate:"required" gorm:"not null"`
	RunningJobs *uint64        `json:"runningJobs,omitempty" validate:"optional" gorm:"default:0"`
	Providers   []ProviderInfo `json:"providers" validate:"required" gorm:"serializer:json;not null"`
} // @name RunnerMetadata

type RunnerState string

var (
	RunnerStateRunning      RunnerState = "running"
	RunnerStateUnresponsive RunnerState = "unresponsive"
)
