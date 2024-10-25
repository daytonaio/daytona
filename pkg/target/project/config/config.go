// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/target/project/buildconfig"
)

type ProjectConfig struct {
	Name                string                   `json:"name" validate:"required"`
	Image               string                   `json:"image" validate:"required"`
	User                string                   `json:"user" validate:"required"`
	BuildConfig         *buildconfig.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	RepositoryUrl       string                   `json:"repositoryUrl" validate:"required"`
	EnvVars             map[string]string        `json:"envVars" validate:"required"`
	IsDefault           bool                     `json:"default" validate:"required"`
	Prebuilds           []*PrebuildConfig        `json:"prebuilds" validate:"optional"`
	GitProviderConfigId *string                  `json:"gitProviderConfigId" validate:"optional"`
} // @name ProjectConfig

func (pc *ProjectConfig) SetPrebuild(p *PrebuildConfig) error {
	newPrebuild := PrebuildConfig{
		Id:             p.Id,
		Branch:         p.Branch,
		CommitInterval: p.CommitInterval,
		TriggerFiles:   p.TriggerFiles,
		Retention:      p.Retention,
	}

	for _, pb := range pc.Prebuilds {
		if pb.Id == p.Id {
			*pb = newPrebuild
			return nil
		}
	}

	pc.Prebuilds = append(pc.Prebuilds, &newPrebuild)
	return nil
}

func (pc *ProjectConfig) FindPrebuild(filter *PrebuildFilter) (*PrebuildConfig, error) {
	for _, pb := range pc.Prebuilds {
		if pb.Match(filter) {
			return pb, nil
		}
	}

	return nil, errors.New("prebuild not found")
}

func (pc *ProjectConfig) ListPrebuilds(filter *PrebuildFilter) ([]*PrebuildConfig, error) {
	if filter == nil {
		return pc.Prebuilds, nil
	}

	prebuilds := []*PrebuildConfig{}

	for _, pb := range pc.Prebuilds {
		if pb.Match(filter) {
			prebuilds = append(prebuilds, pb)
		}
	}

	return prebuilds, nil
}

func (pc *ProjectConfig) RemovePrebuild(id string) error {
	newPrebuilds := []*PrebuildConfig{}

	for _, pb := range pc.Prebuilds {
		if pb.Id != id {
			newPrebuilds = append(newPrebuilds, pb)
		}
	}

	pc.Prebuilds = newPrebuilds
	return nil
}
