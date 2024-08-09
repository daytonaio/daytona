// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

type PrebuildDTO struct {
	ProjectConfigName string   `json:"projectConfigName"`
	Branch            string   `json:"branch"`
	CommitInterval    *int     `json:"commitInterval"`
	TriggerFiles      []string `json:"triggerFiles"`
} // @name PrebuildDTO

// Todo - use PrebuildDTOs
type CreatePrebuildDTO struct {
	ProjectConfigName string   `json:"projectConfigName"`
	Branch            string   `json:"branch"`
	CommitInterval    *int     `json:"commitInterval"`
	TriggerFiles      []string `json:"triggerFiles"`
	RunAtInit         bool     `json:"runAtInit"`
} // @name CreatePrebuildDTO
