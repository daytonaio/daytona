// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	build_dto "github.com/daytonaio/daytona/pkg/server/builds/dto"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	log "github.com/sirupsen/logrus"
)

func (s *ProjectConfigService) SetPrebuild(projectConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	projectConfig, err := s.Find(&config.ProjectConfigFilter{
		Name: &projectConfigName,
	})
	if err != nil {
		return nil, err
	}

	existingPrebuild, err := projectConfig.FindPrebuild(&config.PrebuildFilter{
		Branch: &createPrebuildDto.Branch,
	})

	if err == nil && createPrebuildDto.Id != nil && *createPrebuildDto.Id != existingPrebuild.Id {
		return nil, fmt.Errorf("prebuild for the specified project config and branch already exists")
	}

	if createPrebuildDto.CommitInterval == nil && len(createPrebuildDto.TriggerFiles) == 0 {
		return nil, fmt.Errorf("either the commit interval or trigger files must be specified")
	}

	gitProvider, gitProviderId, err := s.gitProviderService.GetGitProviderForUrl(projectConfig.RepositoryUrl)
	if err != nil {
		return nil, err
	}

	repository, err := gitProvider.GetRepositoryFromUrl(projectConfig.RepositoryUrl)
	if err != nil {
		return nil, err
	}

	prebuild := &config.PrebuildConfig{
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

	err = projectConfig.SetPrebuild(prebuild)
	if err != nil {
		return nil, err
	}

	// Remember the new webhook ID in case config saving fails
	newWebhookId := ""

	existingWebhookId, err := s.gitProviderService.GetPrebuildWebhook(gitProviderId, repository, s.prebuildWebhookEndpoint)
	if err != nil {
		return nil, err
	}

	if existingWebhookId == nil {
		newWebhookId, err = s.gitProviderService.RegisterPrebuildWebhook(gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			return nil, err
		}
	}

	err = s.configStore.Save(projectConfig)
	if err != nil {
		if newWebhookId != "" {
			err = s.gitProviderService.UnregisterPrebuildWebhook(gitProviderId, repository, newWebhookId)
			if err != nil {
				log.Error(err)
			}
		}

		return nil, err
	}

	return &dto.PrebuildDTO{
		Id:                prebuild.Id,
		ProjectConfigName: projectConfig.Name,
		Branch:            prebuild.Branch,
		CommitInterval:    prebuild.CommitInterval,
		TriggerFiles:      prebuild.TriggerFiles,
		Retention:         prebuild.Retention,
	}, nil
}

func (s *ProjectConfigService) FindPrebuild(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error) {
	pc, err := s.configStore.Find(projectConfigFilter)
	if err != nil {
		return nil, config.ErrProjectConfigNotFound
	}

	prebuild, err := pc.FindPrebuild(prebuildFilter)
	if err != nil {
		return nil, err
	}

	return &dto.PrebuildDTO{
		Id:                prebuild.Id,
		ProjectConfigName: pc.Name,
		Branch:            prebuild.Branch,
		CommitInterval:    prebuild.CommitInterval,
		TriggerFiles:      prebuild.TriggerFiles,
		Retention:         prebuild.Retention,
	}, nil
}

func (s *ProjectConfigService) ListPrebuilds(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	var result []*dto.PrebuildDTO
	pcs, err := s.configStore.List(projectConfigFilter)
	if err != nil {
		return nil, config.ErrProjectConfigNotFound
	}

	for _, pc := range pcs {
		for _, prebuild := range pc.Prebuilds {
			result = append(result, &dto.PrebuildDTO{
				Id:                prebuild.Id,
				ProjectConfigName: pc.Name,
				Branch:            prebuild.Branch,
				CommitInterval:    prebuild.CommitInterval,
				TriggerFiles:      prebuild.TriggerFiles,
				Retention:         prebuild.Retention,
			})
		}
	}

	return result, nil
}

func (s *ProjectConfigService) DeletePrebuild(projectConfigName string, id string, force bool) []error {
	projectConfig, err := s.Find(&config.ProjectConfigFilter{
		Name: &projectConfigName,
	})
	if err != nil {
		return []error{err}
	}

	// Get all prebuilds for this project config's repository URL and
	// if this is the last prebuild, unregister the Git provider webhook
	prebuilds, err := s.ListPrebuilds(&config.ProjectConfigFilter{
		Url: &projectConfig.RepositoryUrl,
	}, nil)
	if err != nil {
		return []error{err}
	}

	if len(prebuilds) == 1 {
		gitProvider, gitProviderId, err := s.gitProviderService.GetGitProviderForUrl(projectConfig.RepositoryUrl)
		if err != nil {
			return []error{err}
		}

		repository, err := gitProvider.GetRepositoryFromUrl(projectConfig.RepositoryUrl)
		if err != nil {
			return []error{err}
		}

		existingWebhookId, err := s.gitProviderService.GetPrebuildWebhook(gitProviderId, repository, s.prebuildWebhookEndpoint)
		if err != nil {
			if force {
				log.Error(err)
			} else {
				return []error{err}
			}
		}

		if existingWebhookId != nil {
			err = s.gitProviderService.UnregisterPrebuildWebhook(gitProviderId, repository, *existingWebhookId)
			if err != nil {
				if force {
					log.Error(err)
				} else {
					return []error{err}
				}
			}
		}
	}

	errs := s.buildService.MarkForDeletion(&build.Filter{
		PrebuildIds: &[]string{id},
	})
	if len(errs) > 0 {
		if force {
			for _, err := range errs {
				log.Error(err)
			}
		} else {
			return errs
		}
	}

	err = projectConfig.RemovePrebuild(id)
	if err != nil {
		return []error{err}
	}

	err = s.configStore.Save(projectConfig)
	if err != nil {
		return []error{err}
	}

	return nil
}

func (s *ProjectConfigService) ProcessGitEvent(data gitprovider.GitEventData) error {
	var buildsToTrigger []build.Build

	projectConfigs, err := s.List(&config.ProjectConfigFilter{
		Url: &data.Url,
	})
	if err != nil {
		return err
	}

	gitProvider, _, err := s.gitProviderService.GetGitProviderForUrl(data.Url)
	if err != nil {
		return err
	}

	repo, err := gitProvider.GetRepositoryFromUrl(data.Url)
	if err != nil {
		return err
	}

	for _, projectConfig := range projectConfigs {
		prebuild, err := projectConfig.FindPrebuild(&config.PrebuildFilter{
			Branch: &data.Branch,
		})
		if err != nil || prebuild == nil {
			continue
		}

		// Check if the commit's affected files and prebuild config's trigger files have any overlap
		if len(prebuild.TriggerFiles) > 0 {
			if slicesHaveCommonEntry(prebuild.TriggerFiles, data.AffectedFiles) {
				buildsToTrigger = append(buildsToTrigger, build.Build{
					Image:       projectConfig.Image,
					User:        projectConfig.User,
					BuildConfig: projectConfig.BuildConfig,
					Repository:  repo,
					EnvVars:     projectConfig.EnvVars,
					PrebuildId:  prebuild.Id,
				})
				continue
			}
		}

		newestBuild, err := s.buildService.Find(&build.Filter{
			PrebuildIds: &[]string{prebuild.Id},
			GetNewest:   util.Pointer(true),
		})
		if err != nil {
			buildsToTrigger = append(buildsToTrigger, build.Build{
				Image:       projectConfig.Image,
				User:        projectConfig.User,
				BuildConfig: projectConfig.BuildConfig,
				Repository:  repo,
				EnvVars:     projectConfig.EnvVars,
				PrebuildId:  prebuild.Id,
			})
			continue
		}

		commitsRange, err := gitProvider.GetCommitsRange(repo, data.Owner, newestBuild.Repository.Sha, data.Sha)
		if err != nil {
			return err
		}

		// Check if the commit interval has been reached
		if prebuild.CommitInterval != nil && commitsRange >= *prebuild.CommitInterval {
			buildsToTrigger = append(buildsToTrigger, build.Build{
				Image:       projectConfig.Image,
				User:        projectConfig.User,
				BuildConfig: projectConfig.BuildConfig,
				Repository:  repo,
				EnvVars:     projectConfig.EnvVars,
				PrebuildId:  prebuild.Id,
			})
		}
	}

	for _, build := range buildsToTrigger {
		createBuildDto := build_dto.BuildCreationData{
			Image:       build.Image,
			User:        build.User,
			BuildConfig: build.BuildConfig,
			Repository:  build.Repository,
			EnvVars:     build.EnvVars,
			PrebuildId:  build.PrebuildId,
		}

		_, err = s.buildService.Create(createBuildDto)
		if err != nil {
			return err
		}
	}

	return nil
}

// Marks the [retention] oldest published builds for deletion for each prebuild
func (s *ProjectConfigService) EnforceRetentionPolicy() error {
	prebuilds, err := s.ListPrebuilds(nil, nil)
	if err != nil {
		return err
	}

	builds, err := s.buildService.List(&build.Filter{
		States: &[]build.BuildState{build.BuildStatePublished},
	})
	if err != nil {
		return err
	}

	buildMap := make(map[string][]build.Build)

	// Group builds by their prebuildId
	for _, b := range builds {
		buildMap[b.PrebuildId] = append(buildMap[b.PrebuildId], *b)
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
				errs := s.buildService.MarkForDeletion(&build.Filter{
					Id: &associatedBuilds[i].Id,
				})
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

func (s *ProjectConfigService) StartRetentionPoller() error {
	scheduler := build.NewCronScheduler()

	err := scheduler.AddFunc(build.DEFAULT_POLL_INTERVAL, func() {
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
