// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
)

type IRunnerService interface {
	List(ctx context.Context) ([]*RunnerDTO, error)
	Find(ctx context.Context, runnerId string) (*RunnerDTO, error)
	Create(ctx context.Context, req CreateRunnerDTO) (*RunnerDTO, error)
	Delete(ctx context.Context, runnerId string) error

	UpdateMetadata(ctx context.Context, runnerId string, metadata *models.RunnerMetadata) error
	UpdateJobState(ctx context.Context, jobId string, req UpdateJobStateDTO) error
	ListRunnerJobs(ctx context.Context, runnerId string) ([]*models.Job, error)

	ListProviders(ctx context.Context, runnerId *string) ([]models.ProviderInfo, error)
	ListProvidersForInstall(ctx context.Context) ([]ProviderDTO, error)
	InstallProvider(ctx context.Context, runnerId string, providerDto InstallProviderDTO) error
	UninstallProvider(ctx context.Context, runnerId string, providerName string) error

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

type ProviderDTO struct {
	Name    string  `json:"name" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
	Version string  `json:"version" validate:"required"`
	Latest  bool    `json:"latest" validate:"required"`
} // @name ProviderDTO

type InstallProviderDTO struct {
	Name    string  `json:"name" validate:"required"`
	Version *string `json:"version" validate:"optional"`
} // @name InstallProviderDTO

var (
	ErrRunnerAlreadyExists = errors.New("runner already exists")
)
