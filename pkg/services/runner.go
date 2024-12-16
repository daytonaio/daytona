// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/os"
)

type IRunnerService interface {
	RegisterRunner(ctx context.Context, req RegisterRunnerDTO) (*RunnerDTO, error)
	GetRunner(ctx context.Context, runnerId string) (*RunnerDTO, error)
	ListRunners(ctx context.Context) ([]*RunnerDTO, error)
	ListRunnerJobs(ctx context.Context, runnerId string) ([]*models.Job, error)
	UpdateJobState(ctx context.Context, jobId string, req UpdateJobStateDTO) error
	SetRunnerMetadata(ctx context.Context, runnerId string, metadata *models.RunnerMetadata) error
	RemoveRunner(ctx context.Context, runnerId string) error

	ListProviders(ctx context.Context, runnerId *string) ([]models.ProviderInfo, error)
	InstallProvider(ctx context.Context, runnerId string, providerMetadata InstallProviderDTO) error
	UninstallProvider(ctx context.Context, runnerId string, providerName string) error
	UpdateProvider(ctx context.Context, runnerId string, providerName string, downloadUrls DownloadUrls) error
}

type RunnerDTO struct {
	models.Runner
	State models.RunnerState `json:"state" validate:"required"`
} //	@name	RunnerDTO

type RegisterRunnerDTO struct {
	Id    string `json:"id" validate:"required"`
	Alias string `json:"alias" validate:"required"`
} // @name RegisterRunnerDTO

type RegisterRunnerResultDTO struct {
	models.Runner
	ApiKey string `json:"apiKey" validate:"required"`
} // @name RegisterRunnerResultDTO

type UpdateJobStateDTO struct {
	State        models.JobState `json:"state" validate:"required"`
	ErrorMessage *string         `json:"errorMessage,omitempty" validate:"optional"`
} // @name UpdateJobState

type InstallProviderDTO struct {
	Name         string       `json:"name" validate:"required"`
	DownloadUrls DownloadUrls `json:"downloadUrls" validate:"required"`
} // @name InstallProviderDTO

type DownloadUrls map[os.OperatingSystem]string // @name DownloadUrls

var (
	ErrRunnerAlreadyExists = errors.New("runner already exists")
)
