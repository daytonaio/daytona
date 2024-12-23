// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type WorkspaceServiceConfig struct {
	WorkspaceStore         stores.WorkspaceStore
	WorkspaceMetadataStore stores.WorkspaceMetadataStore

	FindTarget             func(ctx context.Context, targetId string) (*models.Target, error)
	FindContainerRegistry  func(ctx context.Context, image string, envVars map[string]string) *models.ContainerRegistry
	FindCachedBuild        func(ctx context.Context, w *models.Workspace) (*models.CachedBuild, error)
	GenerateApiKey         func(ctx context.Context, name string) (string, error)
	RevokeApiKey           func(ctx context.Context, name string) error
	ListGitProviderConfigs func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error)
	FindGitProviderConfig  func(ctx context.Context, id string) (*models.GitProviderConfig, error)
	GetLastCommitSha       func(ctx context.Context, repo *gitprovider.GitRepository) (string, error)
	CreateJob              func(ctx context.Context, workspaceId string, runnerId string, action models.JobAction) error
	TrackTelemetryEvent    func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error

	LoggerFactory         logs.ILoggerFactory
	ServerApiUrl          string
	ServerUrl             string
	ServerVersion         string
	DefaultWorkspaceImage string
	DefaultWorkspaceUser  string
}

func NewWorkspaceService(config WorkspaceServiceConfig) services.IWorkspaceService {
	return &WorkspaceService{
		workspaceStore:         config.WorkspaceStore,
		workspaceMetadataStore: config.WorkspaceMetadataStore,

		findTarget:             config.FindTarget,
		findContainerRegistry:  config.FindContainerRegistry,
		findCachedBuild:        config.FindCachedBuild,
		generateApiKey:         config.GenerateApiKey,
		revokeApiKey:           config.RevokeApiKey,
		listGitProviderConfigs: config.ListGitProviderConfigs,
		findGitProviderConfig:  config.FindGitProviderConfig,
		getLastCommitSha:       config.GetLastCommitSha,
		createJob:              config.CreateJob,
		trackTelemetryEvent:    config.TrackTelemetryEvent,

		serverApiUrl:          config.ServerApiUrl,
		serverUrl:             config.ServerUrl,
		serverVersion:         config.ServerVersion,
		defaultWorkspaceImage: config.DefaultWorkspaceImage,
		defaultWorkspaceUser:  config.DefaultWorkspaceUser,
		loggerFactory:         config.LoggerFactory,
	}
}

type WorkspaceService struct {
	workspaceStore         stores.WorkspaceStore
	workspaceMetadataStore stores.WorkspaceMetadataStore

	findTarget             func(ctx context.Context, targetId string) (*models.Target, error)
	findContainerRegistry  func(ctx context.Context, image string, envVars map[string]string) *models.ContainerRegistry
	findCachedBuild        func(ctx context.Context, w *models.Workspace) (*models.CachedBuild, error)
	generateApiKey         func(ctx context.Context, name string) (string, error)
	revokeApiKey           func(ctx context.Context, name string) error
	listGitProviderConfigs func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error)
	findGitProviderConfig  func(ctx context.Context, id string) (*models.GitProviderConfig, error)
	getLastCommitSha       func(ctx context.Context, repo *gitprovider.GitRepository) (string, error)
	createJob              func(ctx context.Context, workspaceId string, runnerId string, action models.JobAction) error
	trackTelemetryEvent    func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error

	serverApiUrl          string
	serverUrl             string
	serverVersion         string
	defaultWorkspaceImage string
	defaultWorkspaceUser  string
	loggerFactory         logs.ILoggerFactory
}

func (s *WorkspaceService) GetWorkspaceLogReader(ctx context.Context, workspaceId string) (io.Reader, error) {
	return s.loggerFactory.CreateLogReader(workspaceId)
}

func (s *WorkspaceService) GetWorkspaceLogWriter(ctx context.Context, workspaceId string) (io.WriteCloser, error) {
	return s.loggerFactory.CreateLogWriter(workspaceId)
}

func (s *WorkspaceService) UpdateWorkspaceProviderMetadata(ctx context.Context, workspaceId, metadata string) error {
	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return err
	}

	w.ProviderMetadata = &metadata
	return s.workspaceStore.Save(ctx, w)
}
