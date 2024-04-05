// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type targetStore interface {
	Find(providerName, targetName string) (*provider.ProviderTarget, error)
}

type WorkspaceServiceConfig struct {
	WorkspaceStore     workspace.Store
	TargetStore        targetStore
	ServerApiUrl       string
	ServerUrl          string
	Provisioner        provisioner.Provisioner
	ApiKeyService      apikeys.ApiKeyService
	NewWorkspaceLogger func(workspaceId string) logger.Logger
	NewProjectLogger   func(workspaceId, projectName string) logger.Logger
}

func NewWorkspaceService(config WorkspaceServiceConfig) *WorkspaceService {
	return &WorkspaceService{
		workspaceStore:     config.WorkspaceStore,
		targetStore:        config.TargetStore,
		serverApiUrl:       config.ServerApiUrl,
		serverUrl:          config.ServerUrl,
		provisioner:        config.Provisioner,
		newWorkspaceLogger: config.NewWorkspaceLogger,
		newProjectLogger:   config.NewProjectLogger,
		apiKeyService:      config.ApiKeyService,
	}
}

type WorkspaceService struct {
	workspaceStore     workspace.Store
	targetStore        targetStore
	provisioner        provisioner.Provisioner
	apiKeyService      apikeys.ApiKeyService
	serverApiUrl       string
	serverUrl          string
	newWorkspaceLogger func(workspaceId string) logger.Logger
	newProjectLogger   func(workspaceId, projectName string) logger.Logger
}

func (s *WorkspaceService) parseTargetId(targetId string) (string, string, error) {
	split := strings.Split(targetId, "/")
	if len(split) != 2 {
		return "", "", errors.New("invalid targetId")
	}

	return split[0], split[1], nil
}

func (s *WorkspaceService) toTargetId(providerName, targetName string) string {
	return fmt.Sprintf("%s/%s", providerName, targetName)
}
