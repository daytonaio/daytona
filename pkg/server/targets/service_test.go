// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets_test

import (
	"context"
	"testing"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/server/targetconfigs"
	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"

var tg = &models.Target{
	Id:     "test",
	Name:   "test",
	ApiKey: "test",
	ProviderInfo: models.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var createTargetDTO = dto.CreateTargetDTO{
	Name:             "test",
	Id:               "test",
	TargetConfigName: "test",
}

var tc = models.TargetConfig{
	Name: "test",
	ProviderInfo: models.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var targetInfo = models.TargetInfo{
	Name:             createTargetDTO.Name,
	ProviderMetadata: "provider-metadata-test",
}

func TestTargetService(t *testing.T) {
	tg.EnvVars = targets.GetTargetEnvVars(tg, targets.TargetEnvVarParams{
		ApiUrl:    serverApiUrl,
		ServerUrl: serverUrl,
		ClientId:  "test-client-id",
	}, false)

	ctx := context.Background()
	ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, "test-client-id")

	targetStore := t_targets.NewInMemoryTargetStore()
	targetConfigStore := t_targetconfigs.NewInMemoryTargetConfigStore()

	err := targetConfigStore.Save(&tc)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()
	mockProvisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:       targetStore,
		TargetConfigStore: targetConfigStore,
		ServerApiUrl:      serverApiUrl,
		ServerUrl:         serverUrl,
		ApiKeyService:     apiKeyService,
		Provisioner:       mockProvisioner,
		LoggerFactory:     logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
	})

	t.Run("CreateTarget", func(t *testing.T) {
		mockProvisioner.On("CreateTarget", tg).Return(nil)
		mockProvisioner.On("StartTarget", tg).Return(nil)

		apiKeyService.On("Generate", models.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)

		target, err := service.CreateTarget(ctx, createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		// Must be true after creation
		tg.IsDefault = true

		targetEquals(t, tg, target)

		tg.EnvVars = nil
		tg.ApiKey = ""
	})

	t.Run("CreateTarget fails when target already exists", func(t *testing.T) {
		_, err := service.CreateTarget(ctx, createTargetDTO)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetAlreadyExists, err)
	})

	t.Run("GetTarget", func(t *testing.T) {
		mockProvisioner.On("GetTargetInfo", mock.Anything, tg).Return(&targetInfo, nil)

		target, err := service.GetTarget(ctx, &targets.TargetFilter{IdOrName: &createTargetDTO.Id}, true)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetDtoEquals(t, createTargetDTO, *target, targetInfo, true)
	})

	t.Run("GetTarget fails when target not found", func(t *testing.T) {
		_, err := service.GetTarget(ctx, &targets.TargetFilter{IdOrName: util.Pointer("invalid-id")}, true)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ListTargets", func(t *testing.T) {
		verbose := false
		targets, err := service.ListTargets(ctx, nil, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, verbose)
	})

	t.Run("ListTargets - verbose", func(t *testing.T) {
		verbose := true

		targets, err := service.ListTargets(ctx, nil, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, verbose)
	})

	t.Run("StartTarget", func(t *testing.T) {
		tg.EnvVars = targets.GetTargetEnvVars(tg, targets.TargetEnvVarParams{
			ApiUrl:    serverApiUrl,
			ServerUrl: serverUrl,
			ClientId:  "test-client-id",
		}, false)

		err := service.StartTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		tg.EnvVars = nil
	})

	t.Run("StopTarget", func(t *testing.T) {
		mockProvisioner.On("StopTarget", tg).Return(nil)

		err := service.StopTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		mockProvisioner.On("DestroyTarget", tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(ctx, &targets.TargetFilter{IdOrName: &createTargetDTO.Id}, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ForceRemoveTarget", func(t *testing.T) {
		err := targetStore.Save(tg)
		require.Nil(t, err)

		mockProvisioner.On("DestroyTarget", tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(ctx, &targets.TargetFilter{IdOrName: &createTargetDTO.Id}, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("CreateTarget fails name validation", func(t *testing.T) {
		invalidTargetRequest := createTargetDTO
		invalidTargetRequest.Name = "invalid name"

		_, err := service.CreateTarget(ctx, invalidTargetRequest)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrInvalidTargetName, err)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		mockProvisioner.AssertExpectations(t)
	})
}

func targetEquals(t *testing.T, t1, t2 *models.Target) {
	t.Helper()

	require.Equal(t, t1.Id, t2.Id)
	require.Equal(t, t1.Name, t2.Name)
	require.Equal(t, t1.ProviderInfo, t2.ProviderInfo)
	require.Equal(t, t1.Options, t2.Options)
	require.Equal(t, t1.IsDefault, t2.IsDefault)
}

func targetDtoEquals(t *testing.T, req dto.CreateTargetDTO, target dto.TargetDTO, targetInfo models.TargetInfo, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, tc.ProviderInfo, target.ProviderInfo)
	require.Equal(t, tc.Options, target.Options)

	if verbose {
		require.Equal(t, target.Info.Name, targetInfo.Name)
		require.Equal(t, target.Info.ProviderMetadata, targetInfo.ProviderMetadata)
	} else {
		require.Nil(t, target.Info)
	}
}
