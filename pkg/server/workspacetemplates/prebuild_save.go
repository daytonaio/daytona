// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceTemplateService) SavePrebuild(ctx context.Context, workspaceTemplateName string, createPrebuildDto services.CreatePrebuildDTO) (*services.PrebuildDTO, error) {
	workspaceTemplate, err := s.Find(ctx, &stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return nil, s.handleSavePrebuildError(ctx, nil, err)
	}

	existingPrebuild, _ := workspaceTemplate.FindPrebuild(&models.MatchParams{
		Branch: &createPrebuildDto.Branch,
	})

	if existingPrebuild != nil && createPrebuildDto.Id == nil {
		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, errors.New("prebuild for the specified workspace template and branch already exists"))
	}

	if createPrebuildDto.CommitInterval == nil && len(createPrebuildDto.TriggerFiles) == 0 {
		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, errors.New("either the commit interval or trigger files must be specified"))
	}

	repository, gitProviderId, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl)
	if err != nil {
		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
	}

	prebuild := &models.PrebuildConfig{
		Branch:         createPrebuildDto.Branch,
		CommitInterval: createPrebuildDto.CommitInterval,
		TriggerFiles:   createPrebuildDto.TriggerFiles,
		Retention:      createPrebuildDto.Retention,
	}

	if createPrebuildDto.Id != nil {
		prebuild.Id = *createPrebuildDto.Id
	} else {
		err = prebuild.GenerateId()
		if err != nil {
			return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
		}
	}

	err = workspaceTemplate.SetPrebuild(prebuild)
	if err != nil {
		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
	}

	// Remember the new webhook ID in case config saving fails
	newWebhookId := ""

	existingWebhookId, err := s.findPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
	if err != nil {
		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
	}

	if existingWebhookId == nil {
		newWebhookId, err = s.registerPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
		}
	}

	err = s.templateStore.Save(ctx, workspaceTemplate)
	if err != nil {
		if newWebhookId != "" {
			err = s.unregisterPrebuildWebhook(ctx, gitProviderId, repository, newWebhookId)
			if err != nil {
				log.Error(err)
			}
		}

		return nil, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
	}

	return &services.PrebuildDTO{
		Id:                    prebuild.Id,
		WorkspaceTemplateName: workspaceTemplate.Name,
		Branch:                prebuild.Branch,
		CommitInterval:        prebuild.CommitInterval,
		TriggerFiles:          prebuild.TriggerFiles,
		Retention:             prebuild.Retention,
	}, s.handleSavePrebuildError(ctx, workspaceTemplate, err)
}

func (s *WorkspaceTemplateService) handleSavePrebuildError(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceTemplateEventPrebuildSaved
	if err != nil {
		eventName = telemetry.WorkspaceTemplateEventPrebuildSaveFailed
	}
	event := telemetry.NewWorkspaceTemplateEvent(eventName, workspaceTemplate, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
