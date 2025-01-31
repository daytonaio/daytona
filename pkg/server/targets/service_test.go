// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets_test

import (
	"context"
	"testing"

	"github.com/daytonaio/daytona/internal/testing/job"
	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/server/targetconfigs"
	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"

var tc = models.TargetConfig{
	Id:   "test",
	Name: "test",
	ProviderInfo: models.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
	Deleted: false,
}

var tg = &models.Target{
	Id:             "test",
	Name:           "test",
	ApiKey:         "test",
	TargetConfigId: tc.Id,
	TargetConfig:   tc,
}

var createTargetDTO = services.CreateTargetDTO{
	Name:           "test",
	Id:             "test",
	TargetConfigId: "test",
}

func TestTargetService(t *testing.T) {
	tg.EnvVars = targets.GetTargetEnvVars(tg, targets.TargetEnvVarParams{
		ApiUrl:    serverApiUrl,
		ServerUrl: serverUrl,
		ClientId:  "test-client-id",
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, "test-client-id")

	jobStore := job.NewInMemoryJobStore()

	targetStore := t_targets.NewInMemoryTargetStore(jobStore)
	targetMetadataStore := t_targets.NewInMemoryTargetMetadataStore()
	targetConfigStore := t_targetconfigs.NewInMemoryTargetConfigStore()

	err := targetConfigStore.Save(ctx, &tc)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()

	tgLogsDir := t.TempDir()

	service := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:         targetStore,
		TargetMetadataStore: targetMetadataStore,
		FindTargetConfig: func(ctx context.Context, name string) (*models.TargetConfig, error) {
			return targetConfigStore.Find(ctx, name, false)
		},
		CreateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Create(models.ApiKeyTypeTarget, name)
		},
		DeleteApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Delete(name)
		},
		ServerApiUrl:  serverApiUrl,
		ServerUrl:     serverUrl,
		LoggerFactory: logs.NewLoggerFactory(logs.LoggerFactoryConfig{LogsDir: tgLogsDir}),
		TrackTelemetryEvent: func(event telemetry.Event, clientId string) error {
			return nil
		},
		CreateJob: func(ctx context.Context, targetId, runnerId string, action models.JobAction) error {
			return jobStore.Save(ctx, &models.Job{
				Id:           targetId,
				ResourceId:   targetId,
				RunnerId:     util.Pointer(runnerId),
				ResourceType: models.ResourceTypeTarget,
				Action:       action,
				State:        models.JobStateSuccess,
			})
		},
	})

	t.Run("CreateTarget", func(t *testing.T) {
		apiKeyService.On("Create", models.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)

		target, err := service.Create(ctx, createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetEquals(t, tg, target)

		tg.EnvVars = nil
		tg.ApiKey = ""
	})

	t.Run("CreateTarget fails when target already exists", func(t *testing.T) {
		_, err := service.Create(ctx, createTargetDTO)
		require.NotNil(t, err)
		require.Equal(t, services.ErrTargetAlreadyExists, err)
	})

	t.Run("FindTarget", func(t *testing.T) {
		target, err := service.Find(ctx, &stores.TargetFilter{IdOrName: &createTargetDTO.Id}, services.TargetRetrievalParams{})

		require.Nil(t, err)
		require.NotNil(t, target)

		targetDtoEquals(t, createTargetDTO, *target)
	})

	t.Run("FindTarget fails when target not found", func(t *testing.T) {
		_, err := service.Find(ctx, &stores.TargetFilter{IdOrName: util.Pointer("invalid-id")}, services.TargetRetrievalParams{})
		require.NotNil(t, err)
		require.Equal(t, stores.ErrTargetNotFound, err)
	})

	t.Run("ListTargets", func(t *testing.T) {
		targets, err := service.List(ctx, nil, services.TargetRetrievalParams{})

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target)
	})

	t.Run("StartTarget", func(t *testing.T) {
		tg.EnvVars = targets.GetTargetEnvVars(tg, targets.TargetEnvVarParams{
			ApiUrl:    serverApiUrl,
			ServerUrl: serverUrl,
			ClientId:  "test-client-id",
		})

		err := service.Start(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		tg.EnvVars = nil
	})

	t.Run("StopTarget", func(t *testing.T) {
		err := service.Stop(ctx, createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("UpdateTargetMetadata", func(t *testing.T) {
		err := targetStore.Save(ctx, tg)
		require.Nil(t, err)

		_, err = service.UpdateMetadata(context.TODO(), tg.Id, &models.TargetMetadata{
			Uptime: 10,
		})
		require.Nil(t, err)
	})

	t.Run("DeleteTarget", func(t *testing.T) {
		apiKeyService.On("Delete", mock.Anything).Return(nil)

		err := service.Delete(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.Find(ctx, &stores.TargetFilter{IdOrName: &createTargetDTO.Id}, services.TargetRetrievalParams{})
		require.Equal(t, services.ErrTargetDeleted, err)
	})

	t.Run("ForceDeleteTarget", func(t *testing.T) {
		err := targetStore.Save(ctx, tg)
		require.Nil(t, err)

		apiKeyService.On("Delete", mock.Anything).Return(nil)

		err = service.ForceDelete(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.Find(ctx, &stores.TargetFilter{IdOrName: &createTargetDTO.Id}, services.TargetRetrievalParams{})
		require.Equal(t, services.ErrTargetDeleted, err)
	})

	t.Run("CreateTarget fails name validation", func(t *testing.T) {
		invalidTargetRequest := createTargetDTO
		invalidTargetRequest.Id = "some-id"
		invalidTargetRequest.Name = "invalid name"

		_, err := service.Create(ctx, invalidTargetRequest)
		require.NotNil(t, err)
		require.Equal(t, services.ErrInvalidTargetName, err)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
	})
}

func targetEquals(t *testing.T, t1, t2 *models.Target) {
	t.Helper()

	require.Equal(t, t1.Id, t2.Id)
	require.Equal(t, t1.Name, t2.Name)
	require.Equal(t, t1.TargetConfig.ProviderInfo, t2.TargetConfig.ProviderInfo)
	require.Equal(t, t1.TargetConfig.Options, t2.TargetConfig.Options)
	require.Equal(t, t1.IsDefault, t2.IsDefault)
}

func targetDtoEquals(t *testing.T, req services.CreateTargetDTO, target services.TargetDTO) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, tc.ProviderInfo, target.TargetConfig.ProviderInfo)
	require.Equal(t, tc.Options, target.TargetConfig.Options)
}
