// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/docker/docker/pkg/stringid"
)

type WorkspaceTemplate struct {
	Name                string            `json:"name" validate:"required" gorm:"primaryKey"`
	Image               string            `json:"image" validate:"required" gorm:"not null"`
	User                string            `json:"user" validate:"required" gorm:"not null"`
	BuildConfig         *BuildConfig      `json:"buildConfig,omitempty" validate:"optional" gorm:"serializer:json"`
	RepositoryUrl       string            `json:"repositoryUrl" validate:"required" gorm:"not null"`
	EnvVars             map[string]string `json:"envVars" validate:"required" gorm:"serializer:json;not null"`
	IsDefault           bool              `json:"default" validate:"required" gorm:"not null"`
	Prebuilds           []*PrebuildConfig `json:"prebuilds" validate:"optional" gorm:"serializer:json"`
	GitProviderConfigId *string           `json:"gitProviderConfigId" validate:"optional"`
} // @name WorkspaceTemplate

func (wt *WorkspaceTemplate) SetPrebuild(p *PrebuildConfig) error {
	newPrebuild := PrebuildConfig{
		Id:             p.Id,
		Branch:         p.Branch,
		CommitInterval: p.CommitInterval,
		TriggerFiles:   p.TriggerFiles,
		Retention:      p.Retention,
	}

	for _, pb := range wt.Prebuilds {
		if pb.Id == p.Id {
			*pb = newPrebuild
			return nil
		}
	}

	wt.Prebuilds = append(wt.Prebuilds, &newPrebuild)
	return nil
}

func (wt *WorkspaceTemplate) FindPrebuild(filter *MatchParams) (*PrebuildConfig, error) {
	for _, pb := range wt.Prebuilds {
		if pb.Match(filter) {
			return pb, nil
		}
	}

	return nil, errors.New("prebuild not found")
}

func (wt *WorkspaceTemplate) ListPrebuilds(filter *MatchParams) ([]*PrebuildConfig, error) {
	if filter == nil {
		return wt.Prebuilds, nil
	}

	prebuilds := []*PrebuildConfig{}

	for _, pb := range wt.Prebuilds {
		if pb.Match(filter) {
			prebuilds = append(prebuilds, pb)
		}
	}

	return prebuilds, nil
}

func (wt *WorkspaceTemplate) RemovePrebuild(id string) error {
	newPrebuilds := []*PrebuildConfig{}

	for _, pb := range wt.Prebuilds {
		if pb.Id != id {
			newPrebuilds = append(newPrebuilds, pb)
		}
	}

	wt.Prebuilds = newPrebuilds
	return nil
}

// PrebuildConfig holds configuration for the prebuild process
type PrebuildConfig struct {
	Id             string   `json:"id" validate:"required" gorm:"not null"`
	Branch         string   `json:"branch" validate:"required" gorm:"not null"`
	CommitInterval *int     `json:"commitInterval" validate:"required" gorm:"not null"`
	TriggerFiles   []string `json:"triggerFiles" validate:"required" gorm:"not null"`
	Retention      int      `json:"retention" validate:"required" gorm:"not null"`
} // @name PrebuildConfig

func (p *PrebuildConfig) GenerateId() error {
	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	p.Id = id
	return nil
}

type MatchParams struct {
	WorkspaceTemplateName *string
	Id                    *string
	Branch                *string
	CommitInterval        *int
	TriggerFiles          *[]string
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
