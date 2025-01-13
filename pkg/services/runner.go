// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/os"
)

type IRunnerService interface {
	ListRunners(ctx context.Context) ([]*RunnerDTO, error)
	ListRunnerJobs(ctx context.Context, runnerId string) ([]*models.Job, error)
	FindRunner(ctx context.Context, runnerId string) (*RunnerDTO, error)
	CreateRunner(ctx context.Context, req CreateRunnerDTO) (*RunnerDTO, error)
	UpdateJobState(ctx context.Context, jobId string, req UpdateJobStateDTO) error
	UpdateRunnerMetadata(ctx context.Context, runnerId string, metadata *models.RunnerMetadata) error
	DeleteRunner(ctx context.Context, runnerId string) error

	ListProviders(ctx context.Context, runnerId *string) ([]models.ProviderInfo, error)
	InstallProvider(ctx context.Context, runnerId string, providerDto InstallProviderDTO) error
	UninstallProvider(ctx context.Context, runnerId string, providerName string) error
	UpdateProvider(ctx context.Context, runnerId string, providerName string, providerDto UpdateProviderDTO) error

	GetRunnerLogReader(ctx context.Context, runnerId string) (io.Reader, error)
	GetRunnerLogWriter(ctx context.Context, runnerId string) (io.WriteCloser, error)
}

type RunnerDTO struct {
	models.Runner
	State models.ResourceState `json:"state" validate:"required"`
} //	@name	RunnerDTO

type CreateRunnerDTO struct {
	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
} // @name CreateRunnerDTO

type CreateRunnerResultDTO struct {
	models.Runner
	ApiKey string `json:"apiKey" validate:"required"`
} // @name CreateRunnerResultDTO

type UpdateJobStateDTO struct {
	State        models.JobState `json:"state" validate:"required"`
	ErrorMessage *string         `json:"errorMessage,omitempty" validate:"optional"`
} // @name UpdateJobState

type InstallProviderDTO struct {
	Name         string       `json:"name" validate:"required"`
	DownloadUrls DownloadUrls `json:"downloadUrls" validate:"required"`
	Version      string       `json:"version" validate:"required"`
} // @name InstallProviderDTO

type UpdateProviderDTO struct {
	DownloadUrls DownloadUrls `json:"downloadUrls" validate:"required"`
	Version      string       `json:"version" validate:"required"`
} // @name UpdateProviderDTO

type DownloadUrls map[os.OperatingSystem]string // @name DownloadUrls

var (
	ErrRunnerAlreadyExists = errors.New("runner already exists")
)
