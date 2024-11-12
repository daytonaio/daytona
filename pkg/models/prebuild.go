// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

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

type MatchParams struct {
	WorkspaceConfigName *string
	Id                  *string
	Branch              *string
	CommitInterval      *int
	TriggerFiles        *[]string
}

func (p *PrebuildConfig) Match(params *MatchParams) bool {
	if params.Id != nil && *params.Id != p.Id {
		return false
	}

	if params.Branch != nil && *params.Branch != p.Branch {
		return false
	}

	if params.CommitInterval != nil && p.CommitInterval != nil && *params.CommitInterval != *p.CommitInterval {
		return false
	}

	if params.TriggerFiles != nil {
		// Sort the trigger files before checking if same
		sort.Strings(p.TriggerFiles)
		sort.Strings(*params.TriggerFiles)
		triggerFilesJson, err := json.Marshal(p.TriggerFiles)
		if err != nil {
			return false
		}
		filterFilesJson, err := json.Marshal(*params.TriggerFiles)
		if err != nil {
			return false
		}
		if string(triggerFilesJson) != string(filterFilesJson) {
			return false
		}
	}

	return true
}
