// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package build

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type BuildState string

const (
	BuildStatePendingRun    BuildState = "pending-run"
	BuildStateRunning       BuildState = "running"
	BuildStateError         BuildState = "error"
	BuildStateSuccess       BuildState = "success"
	BuildStatePublished     BuildState = "published"
	BuildStatePendingDelete BuildState = "pending-delete"
	BuildStateDeleting      BuildState = "deleting"
)

type Build struct {
	Id          string                     `json:"id" validate:"required"`
	State       BuildState                 `json:"state" validate:"required"`
	Image       string                     `json:"image" validate:"required"`
	User        string                     `json:"user" validate:"required"`
	BuildConfig *buildconfig.BuildConfig   `json:"buildConfig" validate:"optional"`
	Repository  *gitprovider.GitRepository `json:"repository" validate:"optional"`
	EnvVars     map[string]string          `json:"envVars" validate:"required"`
	PrebuildId  string                     `json:"prebuildId" validate:"required"`
	CreatedAt   time.Time                  `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time                  `json:"updatedAt" validate:"required"`
} // @name Build

func (b *Build) Compare(other *Build) (bool, error) {
	if b.BuildConfig != nil && *b.BuildConfig == (buildconfig.BuildConfig{}) {
		buildHash, err := b.getBuildHashWithoutBuildConfig()
		if err != nil {
			return false, err
		}
		otherHash, err := other.getBuildHashWithoutBuildConfig()
		if err != nil {
			return false, err
		}
		return buildHash == otherHash, nil
	}

	buildHash, err := b.GetBuildHash()
	if err != nil {
		return false, err
	}

	otherHash, err := other.GetBuildHash()
	if err != nil {
		return false, err
	}

	return buildHash == otherHash, nil
}

// GetBuildHash returns a SHA-256 hash of the build's configuration, repository branch and environment variables.
func (b *Build) GetBuildHash() (string, error) {
	var buildJson []byte
	var err error
	if b.BuildConfig != nil && b.BuildConfig.Devcontainer != nil {
		buildJson, err = json.Marshal(b.BuildConfig.Devcontainer)
		if err != nil {
			return "", err
		}
	}
	envVarsJson, err := json.Marshal(b.EnvVars)
	if err != nil {
		return "", err
	}

	data := string(buildJson) + b.Repository.Url + b.Repository.Branch + string(envVarsJson)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr, nil
}

// Helper function used for instances where the build's configuration is automatic
// Returns a SHA-256 hash of only the build's repository branch and environment variables
func (b *Build) getBuildHashWithoutBuildConfig() (string, error) {
	var err error
	envVarsJson, err := json.Marshal(b.EnvVars)
	if err != nil {
		return "", err
	}

	data := b.Repository.Branch + b.Repository.Url + string(envVarsJson)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}

func GetCachedBuild(build *Build, builds []*Build) *buildconfig.CachedBuild {
	var cachedBuild *Build

	for _, existingBuild := range builds {
		equal, err := build.Compare(existingBuild)
		if err != nil {
			continue
		}
		if !equal || existingBuild.State != BuildStatePublished {
			continue
		}
		if cachedBuild == nil {
			cachedBuild = existingBuild
			continue
		}
		if existingBuild.CreatedAt.After(cachedBuild.CreatedAt) {
			cachedBuild = existingBuild
		}
	}

	if cachedBuild != nil {
		return &buildconfig.CachedBuild{
			Image: cachedBuild.Image,
			User:  cachedBuild.User,
		}
	}

	return nil
}
