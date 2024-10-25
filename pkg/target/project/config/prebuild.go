// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"sort"

	"github.com/docker/docker/pkg/stringid"
)

// PrebuildConfig holds configuration for the prebuild process
type PrebuildConfig struct {
	Id             string   `json:"id" validate:"required"`
	Branch         string   `json:"branch" validate:"required"`
	CommitInterval *int     `json:"commitInterval" validate:"required"`
	TriggerFiles   []string `json:"triggerFiles" validate:"required"`
	Retention      int      `json:"retention" validate:"required"`
} // @name PrebuildConfig

func (p *PrebuildConfig) GenerateId() error {
	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	p.Id = id
	return nil
}

func (p *PrebuildConfig) Match(filter *PrebuildFilter) bool {
	if filter.Id != nil && *filter.Id != p.Id {
		return false
	}

	if filter.Branch != nil && *filter.Branch != p.Branch {
		return false
	}

	if filter.CommitInterval != nil && p.CommitInterval != nil && *filter.CommitInterval != *p.CommitInterval {
		return false
	}

	if filter.TriggerFiles != nil {
		// Sort the trigger files before checking if same
		sort.Strings(p.TriggerFiles)
		sort.Strings(*filter.TriggerFiles)
		triggerFilesJson, err := json.Marshal(p.TriggerFiles)
		if err != nil {
			return false
		}
		filterFilesJson, err := json.Marshal(*filter.TriggerFiles)
		if err != nil {
			return false
		}
		if string(triggerFilesJson) != string(filterFilesJson) {
			return false
		}
	}

	return true
}
