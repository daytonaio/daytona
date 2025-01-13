// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates

import (
	"context"
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceTemplateService) FindPrebuild(ctx context.Context, workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) (*services.PrebuildDTO, error) {
	wt, err := s.templateStore.Find(ctx, workspaceTemplateFilter)
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

	return &services.PrebuildDTO{
		Id:                    prebuild.Id,
		WorkspaceTemplateName: wt.Name,
		Branch:                prebuild.Branch,
		CommitInterval:        prebuild.CommitInterval,
		TriggerFiles:          prebuild.TriggerFiles,
		Retention:             prebuild.Retention,
	}, nil
}

func (s *WorkspaceTemplateService) ListPrebuilds(ctx context.Context, workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) ([]*services.PrebuildDTO, error) {
	var result []*services.PrebuildDTO
	wts, err := s.templateStore.List(ctx, workspaceTemplateFilter)
	if err != nil {
		return nil, stores.ErrWorkspaceTemplateNotFound
	}

	for _, wt := range wts {
		for _, prebuild := range wt.Prebuilds {
			result = append(result, &services.PrebuildDTO{
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

// TODO: revise build trigger strategy
// We should discuss if the function should throw if the build can not be created or move on to the next one
func (s *WorkspaceTemplateService) ProcessGitEvent(ctx context.Context, data gitprovider.GitEventData) error {
	workspaceTemplates, err := s.List(ctx, &stores.WorkspaceTemplateFilter{
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
func (s *WorkspaceTemplateService) EnforceRetentionPolicy(ctx context.Context) error {
	prebuilds, err := s.ListPrebuilds(ctx, nil, nil)
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
		if b.PrebuildId != nil && *b.PrebuildId != "" {
			buildMap[*b.PrebuildId] = append(buildMap[*b.PrebuildId], b.Build)
		}
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

func (s *WorkspaceTemplateService) StartRetentionPoller(ctx context.Context) error {
	scheduler := scheduler.NewCronScheduler()

	err := scheduler.AddFunc(runner.DEFAULT_JOB_POLL_INTERVAL, func() {
		err := s.EnforceRetentionPolicy(ctx)
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
