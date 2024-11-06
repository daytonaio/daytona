// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets_test

import (
	"context"
	"testing"

	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"

var tg = target.Target{
	Id:   "test",
	Name: "test",
	ProviderInfo: target.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var createTargetDTO = dto.CreateTargetDTO{
	Name:         "test",
	Id:           "test",
	ProviderInfo: tg.ProviderInfo,
	Options:      tg.Options,
}

var targetInfo = target.TargetInfo{
	Name:             createTargetDTO.Name,
	ProviderMetadata: "provider-metadata-test",
}

func TestTargetService(t *testing.T) {
	targetStore := t_targets.NewInMemoryTargetStore()

	apiKeyService := mocks.NewMockApiKeyService()
	provisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:   targetStore,
		ServerApiUrl:  serverApiUrl,
		ServerUrl:     serverUrl,
		ApiKeyService: apiKeyService,
		Provisioner:   provisioner,
		LoggerFactory: logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
	})

	t.Run("CreateTarget", func(t *testing.T) {
		provisioner.On("CreateTarget", &tg).Return(nil)
		provisioner.On("StartTarget", &tg).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)

		target, err := service.CreateTarget(context.TODO(), createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetEquals(t, createTargetDTO, target)
	})

	t.Run("CreateTarget fails when target already exists", func(t *testing.T) {
		_, err := service.CreateTarget(context.TODO(), createTargetDTO)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetAlreadyExists, err)
	})

	t.Run("CreateTarget fails name validation", func(t *testing.T) {
		invalidTargetRequest := createTargetDTO
		invalidTargetRequest.Name = "invalid name"

		_, err := service.CreateTarget(context.TODO(), invalidTargetRequest)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrInvalidTargetName, err)
	})

	t.Run("GetTarget", func(t *testing.T) {
		provisioner.On("GetTargetInfo", context.TODO(), &tg).Return(&targetInfo, nil)

		target, err := service.GetTarget(context.TODO(), &target.TargetFilter{IdOrName: &createTargetDTO.Id}, true)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetDtoEquals(t, createTargetDTO, *target, targetInfo, true)
	})

	t.Run("GetTarget fails when target not found", func(t *testing.T) {
		_, err := service.GetTarget(context.TODO(), &target.TargetFilter{IdOrName: util.Pointer("invalid-id")}, true)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ListTargets", func(t *testing.T) {
		verbose := false
		provisioner.On("GetTargetInfo", context.TODO(), &tg).Return(&targetInfo, nil)

		targets, err := service.ListTargets(context.TODO(), nil, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, verbose)
	})

	t.Run("ListTargets - verbose", func(t *testing.T) {
		verbose := true
		provisioner.On("GetTargetInfo", context.TODO(), &tg).Return(&targetInfo, nil)

		targets, err := service.ListTargets(context.TODO(), nil, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, verbose)
	})

	t.Run("StartTarget", func(t *testing.T) {
		provisioner.On("StartTarget", &tg, &tg).Return(nil)

		err := service.StartTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopTarget", func(t *testing.T) {
		provisioner.On("StopTarget", &tg, &tg).Return(nil)

		err := service.StopTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		provisioner.On("DestroyTarget", &tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(context.TODO(), &target.TargetFilter{IdOrName: &createTargetDTO.Id}, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ForceRemoveTarget", func(t *testing.T) {
		provisioner.On("CreateTarget", &tg).Return(nil)
		provisioner.On("StartTarget", &tg).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)

		_, _ = service.CreateTarget(context.TODO(), createTargetDTO)

		provisioner.On("DestroyTarget", &tg, &tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.ForceRemoveTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(context.TODO(), &target.TargetFilter{IdOrName: &createTargetDTO.Id}, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		provisioner.AssertExpectations(t)
	})
}

func targetEquals(t *testing.T, req dto.CreateTargetDTO, target *target.Target) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, req.ProviderInfo, target.ProviderInfo)
	require.Equal(t, req.Options, target.Options)
}

func targetDtoEquals(t *testing.T, req dto.CreateTargetDTO, target dto.TargetDTO, targetInfo target.TargetInfo, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, req.ProviderInfo, target.ProviderInfo)
	require.Equal(t, req.Options, target.Options)

	if verbose {
		require.Equal(t, target.Info.Name, targetInfo.Name)
		require.Equal(t, target.Info.ProviderMetadata, targetInfo.ProviderMetadata)
	} else {
		require.Nil(t, target.Info)
	}
}
