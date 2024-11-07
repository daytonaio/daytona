// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"
	"regexp"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/config"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"

	log "github.com/sirupsen/logrus"
)

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

func (s *TargetService) CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*target.Target, error) {
	_, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &req.Id})
	if err == nil {
		return nil, ErrTargetAlreadyExists
	}

	tc, err := s.targetConfigStore.Find(&config.TargetConfigFilter{Name: &req.TargetConfigName})
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	// Repo name is taken as the name for target by default
	if !isValidTargetName(req.Name) {
		return nil, ErrInvalidTargetName
	}

	tg := &target.Target{
		Id:           req.Id,
		Name:         req.Name,
		ProviderInfo: tc.ProviderInfo,
		Options:      tc.Options,
	}

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeTarget, tg.Id)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}
	tg.ApiKey = apiKey

	err = s.targetStore.Save(tg)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	targetLogger := s.loggerFactory.CreateTargetLogger(tg.Id, tg.Name, logs.LogSourceServer)
	defer targetLogger.Close()

	targetLogger.Write([]byte(fmt.Sprintf("Creating target %s (%s)\n", tg.Name, tg.Id)))

	tg.EnvVars = target.GetTargetEnvVars(tg, target.TargetEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.provisioner.CreateTarget(tg)
	if err != nil {
		return s.handleCreateError(ctx, tg, err)
	}

	targetLogger.Write([]byte(views.GetPrettyLogLine("Target creation complete")))

	err = s.startTarget(tg, targetLogger)
	if err != nil {
		return s.handleCreateError(ctx, tg, err)
	}

	tg, err = s.handleCreateError(ctx, tg, err)
	if err != nil {
		return nil, err
	}

	err = s.SetDefault(ctx, tg.Id)
	if err != nil {
		return nil, err
	}

	tg.IsDefault = true

	return s.handleCreateError(ctx, tg, err)
}

func (s *TargetService) handleCreateError(ctx context.Context, target *target.Target, err error) (*target.Target, error) {
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
