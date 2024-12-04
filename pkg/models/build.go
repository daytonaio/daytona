// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type Build struct {
	Id              string                     `json:"id" validate:"required" gorm:"primaryKey"`
	Image           *string                    `json:"image" validate:"optional"`
	User            *string                    `json:"user" validate:"optional"`
	ContainerConfig ContainerConfig            `json:"containerConfig" validate:"required" gorm:"serializer:json"`
	BuildConfig     *BuildConfig               `json:"buildConfig" validate:"optional" gorm:"serializer:json"`
	Repository      *gitprovider.GitRepository `json:"repository" validate:"required" gorm:"serializer:json"`
	EnvVars         map[string]string          `json:"envVars" validate:"required" gorm:"serializer:json"`
	LastJob         *Job                       `gorm:"foreignKey:ResourceId;references:Id" validate:"optional"`
	PrebuildId      string                     `json:"prebuildId" validate:"required"`
	CreatedAt       time.Time                  `json:"createdAt" validate:"required"`
	UpdatedAt       time.Time                  `json:"updatedAt" validate:"required"`
} // @name Build

func (w *Build) GetState() ResourceState {
	return getResourceStateFromJob(w.LastJob)
}

type ContainerConfig struct {
	Image string `json:"image" validate:"required"`
	User  string `json:"user" validate:"required"`
} // @name ContainerConfig

func (b *Build) Compare(other *Build) (bool, error) {
	if b.BuildConfig != nil && *b.BuildConfig == (BuildConfig{}) {
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

func GetCachedBuild(build *Build, builds []*Build) *CachedBuild {
	var cachedBuild *Build

	for _, existingBuild := range builds {
		equal, err := build.Compare(existingBuild)
		if err != nil {
			continue
		}
		if !equal || existingBuild.GetState().Name != ResourceStateNameRunSuccessful {
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

	if cachedBuild != nil && cachedBuild.Image != nil && cachedBuild.User != nil {
		return &CachedBuild{
			Image: *cachedBuild.Image,
			User:  *cachedBuild.User,
		}
	}

	return nil
}
