// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
)

type IBuilder interface {
	Build(build models.Build) (string, string, error)
	CleanUp() error
	Publish(build models.Build) error
	GetImageName(build models.Build) (string, error)
}

type Builder struct {
	id                          string
	workspaceDir                string
	image                       string
	containerRegistries         common.ContainerRegistries
	buildImageContainerRegistry *models.ContainerRegistry

	buildImageNamespace   string
	loggerFactory         logs.ILoggerFactory
	defaultWorkspaceImage string
	defaultWorkspaceUser  string
}

func (b *Builder) GetImageName(build models.Build) (string, error) {
	hash, err := build.GetBuildHash()
	if err != nil {
		return "", err
	}
	tagBytes := sha256.Sum256([]byte(fmt.Sprintf("%s%s", hash, build.Repository.Sha)))
	nameBytes := sha256.Sum256([]byte(build.Repository.Url))

	tag := hex.EncodeToString(tagBytes[:])[:16]
	name := hex.EncodeToString(nameBytes[:])[:16]

	namespace := b.buildImageNamespace
	imageName := fmt.Sprintf("%s%s/w-%s:%s", b.buildImageContainerRegistry.Server, namespace, name, tag)

	return imageName, nil
}
