// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func (s *TargetService) CreateTarget(ctx context.Context, req services.CreateTargetDTO) (*models.Target, error) {
	var err error
	ctx, err = s.targetStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.targetStore)

	_, err = s.targetStore.Find(ctx, &stores.TargetFilter{IdOrName: &req.Id})
	if err == nil {
		return nil, services.ErrTargetAlreadyExists
	}

	tc, err := s.findTargetConfig(ctx, req.TargetConfigName)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	// Repo name is taken as the name for target by default
	if !isValidTargetName(req.Name) {
		return s.handleCreateError(ctx, nil, services.ErrInvalidTargetName)
	}

	tg := &models.Target{
		Id:             req.Id,
		Name:           req.Name,
		TargetConfigId: tc.Id,
		TargetConfig:   *tc,
	}

	apiKey, err := s.generateApiKey(ctx, tg.Id)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}
	tg.ApiKey = apiKey

	daytonaTargetEnvVars := GetTargetEnvVars(tg, TargetEnvVarParams{
		ApiUrl:           s.serverApiUrl,
		ServerUrl:        s.serverUrl,
		ServerVersion:    s.serverVersion,
		ClientId:         telemetry.ClientId(ctx),
		TelemetryEnabled: telemetry.TelemetryEnabled(ctx),
	})
	tg.EnvVars = util.MergeEnvVars(daytonaTargetEnvVars, tg.EnvVars)

	err = s.targetStore.Save(ctx, tg)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	err = s.targetMetadataStore.Save(ctx, &models.TargetMetadata{
		TargetId: tg.Id,
		Uptime:   0,
	})
	if err != nil {
		return s.handleCreateError(ctx, tg, err)
	}

	err = s.createJob(ctx, tg.Id, models.JobActionCreate)
	if err != nil {
		return s.handleCreateError(ctx, tg, err)
	}

	err = s.targetStore.CommitTransaction(ctx)
	return s.handleCreateError(ctx, tg, err)
}

func (s *TargetService) HandleSuccessfulCreation(ctx context.Context, targetId string) error {
	return s.SetDefault(ctx, targetId)
}

func (s *TargetService) handleCreateError(ctx context.Context, target *models.Target, err error) (*models.Target, error) {
	if err != nil {
		err = s.targetStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return target, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target)
	event := telemetry.ServerEventTargetCreated
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetCreateError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return target, err
}

func isValidTargetName(name string) bool {
	// The repository name can only contain ASCII letters, digits, and the characters ., -, and _.
	var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Check if the name matches the basic regex
	if !validName.MatchString(name) {
		return false
	}

	// Names starting with a period must have atleast one char appended to it.
	if name == "." || name == "" {
		return false
	}

	return true
}
