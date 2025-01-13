// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceTemplateService) DeletePrebuild(ctx context.Context, workspaceTemplateName string, id string, force bool) []error {
	workspaceTemplate, err := s.Find(ctx, &stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return []error{s.handleDeletePrebuildError(ctx, nil, err)}
	}

	// Get all prebuilds for this workspace template's repository URL and
	// if this is the last prebuild, unregister the Git provider webhook
	prebuilds, err := s.ListPrebuilds(ctx, &stores.WorkspaceTemplateFilter{
		Url: &workspaceTemplate.RepositoryUrl,
	}, nil)
	if err != nil {
		return []error{s.handleDeletePrebuildError(ctx, workspaceTemplate, err)}
	}

	if len(prebuilds) == 1 {
		repository, gitProviderId, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl)
		if err != nil {
			return []error{s.handleDeletePrebuildError(ctx, workspaceTemplate, err)}
		}

		existingWebhookId, err := s.findPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			if force {
				log.Error(s.handleDeletePrebuildError(ctx, workspaceTemplate, err))
			} else {
				return []error{s.handleDeletePrebuildError(ctx, workspaceTemplate, err)}
			}
		}

		if existingWebhookId != nil {
			err = s.unregisterPrebuildWebhook(ctx, gitProviderId, repository, *existingWebhookId)
			if err != nil {
				if force {
					log.Error(s.handleDeletePrebuildError(ctx, workspaceTemplate, err))
				} else {
					return []error{s.handleDeletePrebuildError(ctx, workspaceTemplate, err)}
				}
			}
		}
	}

	errs := s.deleteBuilds(ctx, &id, nil, force)
	if len(errs) > 0 {
		for _, err := range errs {
			err = s.handleDeletePrebuildError(ctx, workspaceTemplate, err)
			if force {
				log.Error(err)
			}
		}
		if !force {
			return errs
		}
	}

	err = workspaceTemplate.RemovePrebuild(id)
	if err != nil {
		return []error{s.handleDeletePrebuildError(ctx, workspaceTemplate, err)}
	}

	err = s.templateStore.Save(ctx, workspaceTemplate)
	err = s.handleDeletePrebuildError(ctx, workspaceTemplate, err)
	if err != nil {
		return []error{err}
	}

	return nil
}

func (s *WorkspaceTemplateService) handleDeletePrebuildError(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	eventName := telemetry.WorkspaceTemplateEventPrebuildDeleted
	if err != nil {
		eventName = telemetry.WorkspaceTemplateEventPrebuildDeletionFailed
	}

	event := telemetry.NewWorkspaceTemplateEvent(eventName, workspaceTemplate, err, nil)
	telemetryError := s.trackTelemetryEvent(event, telemetry.ClientId(ctx))
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
