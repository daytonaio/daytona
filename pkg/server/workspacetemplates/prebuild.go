// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runners"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/server/workspacetemplates/dto"
	"github.com/daytonaio/daytona/pkg/stores"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceTemplateService) SetPrebuild(workspaceTemplateName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	ctx := context.Background()

	workspaceTemplate, err := s.Find(&stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return nil, err
	}

	existingPrebuild, _ := workspaceTemplate.FindPrebuild(&models.MatchParams{
		Branch: &createPrebuildDto.Branch,
	})

	if existingPrebuild != nil && createPrebuildDto.Id == nil {
		return nil, errors.New("prebuild for the specified workspace template and branch already exists")
	}

	if createPrebuildDto.CommitInterval == nil && len(createPrebuildDto.TriggerFiles) == 0 {
		return nil, errors.New("either the commit interval or trigger files must be specified")
	}

	repository, gitProviderId, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl)
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	err = workspaceTemplate.SetPrebuild(prebuild)
	if err != nil {
		return nil, err
	}

	// Remember the new webhook ID in case config saving fails
	newWebhookId := ""

	existingWebhookId, err := s.findPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
	if err != nil {
		return nil, err
	}

	if existingWebhookId == nil {
		newWebhookId, err = s.registerPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			return nil, err
		}
	}

	err = s.templateStore.Save(workspaceTemplate)
	if err != nil {
		if newWebhookId != "" {
			err = s.unregisterPrebuildWebhook(ctx, gitProviderId, repository, newWebhookId)
			if err != nil {
				log.Error(err)
			}
		}

		return nil, err
	}

	return &dto.PrebuildDTO{
		Id:                    prebuild.Id,
		WorkspaceTemplateName: workspaceTemplate.Name,
		Branch:                prebuild.Branch,
		CommitInterval:        prebuild.CommitInterval,
		TriggerFiles:          prebuild.TriggerFiles,
		Retention:             prebuild.Retention,
	}, nil
}

func (s *WorkspaceTemplateService) FindPrebuild(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) (*dto.PrebuildDTO, error) {
	wt, err := s.templateStore.Find(workspaceTemplateFilter)
	if err != nil {
		return nil, stores.ErrWorkspaceTemplateNotFound
	}

	prebuild, err := wt.FindPrebuild(&models.MatchParams{
		Id:                    prebuildFilter.Id,
		Branch:                prebuildFilter.Branch,
		CommitInterval:        prebuildFilter.CommitInterval,
		TriggerFiles:          prebuildFilter.TriggerFiles,
		WorkspaceTemplateName: prebuildFilter.WorkspaceTemplateName,
	})
	if err != nil {
		return nil, err
	}

	return &dto.PrebuildDTO{
		Id:                    prebuild.Id,
		WorkspaceTemplateName: wt.Name,
		Branch:                prebuild.Branch,
		CommitInterval:        prebuild.CommitInterval,
		TriggerFiles:          prebuild.TriggerFiles,
		Retention:             prebuild.Retention,
	}, nil
}

func (s *WorkspaceTemplateService) ListPrebuilds(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	var result []*dto.PrebuildDTO
	wts, err := s.templateStore.List(workspaceTemplateFilter)
	if err != nil {
		return nil, stores.ErrWorkspaceTemplateNotFound
	}

	for _, wt := range wts {
		for _, prebuild := range wt.Prebuilds {
			result = append(result, &dto.PrebuildDTO{
				Id:                    prebuild.Id,
				WorkspaceTemplateName: wt.Name,
				Branch:                prebuild.Branch,
				CommitInterval:        prebuild.CommitInterval,
				TriggerFiles:          prebuild.TriggerFiles,
				Retention:             prebuild.Retention,
			})
		}
	}

	return result, nil
}

func (s *WorkspaceTemplateService) DeletePrebuild(workspaceTemplateName string, id string, force bool) []error {
	ctx := context.Background()

	workspaceTemplate, err := s.Find(&stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return []error{err}
	}

	// Get all prebuilds for this workspace template's repository URL and
	// if this is the last prebuild, unregister the Git provider webhook
	prebuilds, err := s.ListPrebuilds(&stores.WorkspaceTemplateFilter{
		Url: &workspaceTemplate.RepositoryUrl,
	}, nil)
	if err != nil {
		return []error{err}
	}

	if len(prebuilds) == 1 {
		repository, gitProviderId, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl)
		if err != nil {
			return []error{err}
		}

		existingWebhookId, err := s.findPrebuildWebhook(ctx, gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			if force {
				log.Error(err)
			} else {
				return []error{err}
			}
		}

		if existingWebhookId != nil {
			err = s.unregisterPrebuildWebhook(ctx, gitProviderId, repository, *existingWebhookId)
			if err != nil {
				if force {
					log.Error(err)
				} else {
					return []error{err}
				}
			}
		}
	}

	errs := s.deleteBuilds(ctx, &id, nil, force)
	if len(errs) > 0 {
		if force {
			for _, err := range errs {
				log.Error(err)
			}
		} else {
			return errs
		}
	}

	err = workspaceTemplate.RemovePrebuild(id)
	if err != nil {
		return []error{err}
	}

	err = s.templateStore.Save(workspaceTemplate)
	if err != nil {
		return []error{err}
	}

	return nil
}

// TODO: revise build trigger strategy
// We should discuss if the function should throw if the build can not be created or move on to the next one
func (s *WorkspaceTemplateService) ProcessGitEvent(data gitprovider.GitEventData) error {
	ctx := context.Background()

	workspaceTemplates, err := s.List(&stores.WorkspaceTemplateFilter{
		Url: &data.Url,
	})
	if err != nil {
		return err
	}

	repo, _, err := s.getRepositoryContext(ctx, data.Url)
	if err != nil {
		return fmt.Errorf("failed to get repository context: %s", err)
	}

	for _, workspaceTemplate := range workspaceTemplates {
		prebuild, err := workspaceTemplate.FindPrebuild(&models.MatchParams{
			Branch: &data.Branch,
		})
		if err != nil || prebuild == nil {
			continue
		}

		// Check if the commit's affected files and prebuild config's trigger files have any overlap
		if len(prebuild.TriggerFiles) > 0 {
			if slicesHaveCommonEntry(prebuild.TriggerFiles, data.AffectedFiles) {
				err := s.createBuild(ctx, workspaceTemplate, repo, prebuild.Id)
				if err != nil {
					return fmt.Errorf("failed to create build: %s", err)
				}
				continue
			}
		}

		newestBuild, err := s.findNewestBuild(ctx, prebuild.Id)
		if err != nil {
			err := s.createBuild(ctx, workspaceTemplate, repo, prebuild.Id)
			if err != nil {
				return fmt.Errorf("failed to create build: %s", err)
			}
			continue
		}

		commitsRange, err := s.getCommitsRange(ctx, repo, newestBuild.Repository.Sha, data.Sha)
		if err != nil {
			return fmt.Errorf("failed to get commits range: %s", err)
		}

		// Check if the commit interval has been reached
		if prebuild.CommitInterval != nil && commitsRange >= *prebuild.CommitInterval {
			err := s.createBuild(ctx, workspaceTemplate, repo, prebuild.Id)
			if err != nil {
				return fmt.Errorf("failed to create build: %s", err)
			}
		}
	}

	return nil
}

// Marks the [retention] oldest published builds for deletion for each prebuild
func (s *WorkspaceTemplateService) EnforceRetentionPolicy() error {
	ctx := context.Background()

	prebuilds, err := s.ListPrebuilds(nil, nil)
	if err != nil {
		return err
	}

	existingBuilds, err := s.listSuccessfulBuilds(ctx)
	if err != nil {
		return err
	}

	buildMap := make(map[string][]models.Build)

	// Group builds by their prebuildId
	for _, b := range existingBuilds {
		buildMap[b.PrebuildId] = append(buildMap[b.PrebuildId], b.Build)
	}

	for _, prebuild := range prebuilds {

		associatedBuilds := buildMap[prebuild.Id]

		if len(associatedBuilds) > prebuild.Retention {
			// Sort the builds by creation time in ascending order (oldest first)
			sort.Slice(associatedBuilds, func(i, j int) bool {
				return associatedBuilds[i].CreatedAt.Before(associatedBuilds[j].CreatedAt)
			})

			numToDelete := len(associatedBuilds) - prebuild.Retention

			// Mark the oldest builds for deletion
			for i := 0; i < numToDelete; i++ {
				errs := s.deleteBuilds(ctx, &associatedBuilds[i].Id, nil, false)
				if len(errs) > 0 {
					for _, err := range errs {
						log.Error(err)
					}
				}
			}
		}

	}

	return nil
}

func (s *WorkspaceTemplateService) StartRetentionPoller() error {
	scheduler := scheduler.NewCronScheduler()

	err := scheduler.AddFunc(runners.DEFAULT_JOB_POLL_INTERVAL, func() {
		err := s.EnforceRetentionPolicy()
		if err != nil {
			log.Error(err)
		}
	})
	if err != nil {
		return err
	}

	scheduler.Start()
	return nil
}

func slicesHaveCommonEntry(slice1, slice2 []string) bool {
	entryMap := make(map[string]bool)

	for _, entry := range slice1 {
		entryMap[entry] = true
	}

	for _, entry := range slice2 {
		if entryMap[entry] {
			return true
		}
	}

	return false
}
