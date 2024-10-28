// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"errors"
	"fmt"
	"sort"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	build_dto "github.com/daytonaio/daytona/pkg/server/builds/dto"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig/dto"
	"github.com/daytonaio/daytona/pkg/target/workspace/config"
	"github.com/daytonaio/daytona/pkg/target/workspace/containerconfig"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceConfigService) SetPrebuild(workspaceConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	workspaceConfig, err := s.Find(&config.WorkspaceConfigFilter{
		Name: &workspaceConfigName,
	})
	if err != nil {
		return nil, err
	}

	existingPrebuild, _ := workspaceConfig.FindPrebuild(&config.PrebuildFilter{
		Branch: &createPrebuildDto.Branch,
	})

	if existingPrebuild != nil && createPrebuildDto.Id == nil {
		return nil, errors.New("prebuild for the specified workspace config and branch already exists")
	}

	if createPrebuildDto.CommitInterval == nil && len(createPrebuildDto.TriggerFiles) == 0 {
		return nil, errors.New("either the commit interval or trigger files must be specified")
	}

	gitProvider, gitProviderId, err := s.gitProviderService.GetGitProviderForUrl(workspaceConfig.RepositoryUrl)
	if err != nil {
		return nil, err
	}

	repository, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
		Url: workspaceConfig.RepositoryUrl,
	})
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

	err = workspaceConfig.SetPrebuild(prebuild)
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

	err = s.configStore.Save(workspaceConfig)
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
		Id:                  prebuild.Id,
		WorkspaceConfigName: workspaceConfig.Name,
		Branch:              prebuild.Branch,
		CommitInterval:      prebuild.CommitInterval,
		TriggerFiles:        prebuild.TriggerFiles,
		Retention:           prebuild.Retention,
	}, nil
}

func (s *WorkspaceConfigService) FindPrebuild(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error) {
	wc, err := s.configStore.Find(workspaceConfigFilter)
	if err != nil {
		return nil, config.ErrWorkspaceConfigNotFound
	}

	prebuild, err := wc.FindPrebuild(prebuildFilter)
	if err != nil {
		return nil, err
	}

	return &dto.PrebuildDTO{
		Id:                  prebuild.Id,
		WorkspaceConfigName: wc.Name,
		Branch:              prebuild.Branch,
		CommitInterval:      prebuild.CommitInterval,
		TriggerFiles:        prebuild.TriggerFiles,
		Retention:           prebuild.Retention,
	}, nil
}

func (s *WorkspaceConfigService) ListPrebuilds(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	var result []*dto.PrebuildDTO
	wcs, err := s.configStore.List(workspaceConfigFilter)
	if err != nil {
		return nil, config.ErrWorkspaceConfigNotFound
	}

	for _, wc := range wcs {
		for _, prebuild := range wc.Prebuilds {
			result = append(result, &dto.PrebuildDTO{
				Id:                  prebuild.Id,
				WorkspaceConfigName: wc.Name,
				Branch:              prebuild.Branch,
				CommitInterval:      prebuild.CommitInterval,
				TriggerFiles:        prebuild.TriggerFiles,
				Retention:           prebuild.Retention,
			})
		}
	}

	return result, nil
}

func (s *WorkspaceConfigService) DeletePrebuild(workspaceConfigName string, id string, force bool) []error {
	workspaceConfig, err := s.Find(&config.WorkspaceConfigFilter{
		Name: &workspaceConfigName,
	})
	if err != nil {
		return []error{err}
	}

	// Get all prebuilds for this workspace config's repository URL and
	// if this is the last prebuild, unregister the Git provider webhook
	prebuilds, err := s.ListPrebuilds(&config.WorkspaceConfigFilter{
		Url: &workspaceConfig.RepositoryUrl,
	}, nil)
	if err != nil {
		return []error{err}
	}

	if len(prebuilds) == 1 {
		gitProvider, gitProviderId, err := s.gitProviderService.GetGitProviderForUrl(workspaceConfig.RepositoryUrl)
		if err != nil {
			return []error{err}
		}

		repository, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
			Url: workspaceConfig.RepositoryUrl,
		})
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
	}, force)
	if len(errs) > 0 {
		if force {
			for _, err := range errs {
				log.Error(err)
			}
		} else {
			return errs
		}
	}

	err = workspaceConfig.RemovePrebuild(id)
	if err != nil {
		return []error{err}
	}

	err = s.configStore.Save(workspaceConfig)
	if err != nil {
		return []error{err}
	}

	return nil
}

func (s *WorkspaceConfigService) ProcessGitEvent(data gitprovider.GitEventData) error {
	var buildsToTrigger []build.Build

	workspaceConfigs, err := s.List(&config.WorkspaceConfigFilter{
		Url: &data.Url,
	})
	if err != nil {
		return err
	}
	gitProvider, _, err := s.gitProviderService.GetGitProviderForUrl(data.Url)
	if err != nil {
		return fmt.Errorf("failed to get git provider for URL: %s", err)
	}

	repo, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
		Url: data.Url,
	})
	if err != nil {
		return fmt.Errorf("failed to get repository context: %s", err)
	}

	for _, workspaceConfig := range workspaceConfigs {
		prebuild, err := workspaceConfig.FindPrebuild(&config.PrebuildFilter{
			Branch: &data.Branch,
		})
		if err != nil || prebuild == nil {
			continue
		}

		// Check if the commit's affected files and prebuild config's trigger files have any overlap
		if len(prebuild.TriggerFiles) > 0 {
			if slicesHaveCommonEntry(prebuild.TriggerFiles, data.AffectedFiles) {
				buildsToTrigger = append(buildsToTrigger, build.Build{
					ContainerConfig: containerconfig.ContainerConfig{
						Image: workspaceConfig.Image,
						User:  workspaceConfig.User,
					},
					BuildConfig: workspaceConfig.BuildConfig,
					Repository:  repo,
					EnvVars:     workspaceConfig.EnvVars,
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
				ContainerConfig: containerconfig.ContainerConfig{
					Image: workspaceConfig.Image,
					User:  workspaceConfig.User,
				},
				BuildConfig: workspaceConfig.BuildConfig,
				Repository:  repo,
				EnvVars:     workspaceConfig.EnvVars,
				PrebuildId:  prebuild.Id,
			})
			continue
		}

		commitsRange, err := gitProvider.GetCommitsRange(repo, newestBuild.Repository.Sha, data.Sha)
		if err != nil {
			return fmt.Errorf("failed to get commits range: %s", err)
		}

		// Check if the commit interval has been reached
		if prebuild.CommitInterval != nil && commitsRange >= *prebuild.CommitInterval {
			buildsToTrigger = append(buildsToTrigger, build.Build{
				ContainerConfig: containerconfig.ContainerConfig{
					Image: workspaceConfig.Image,
					User:  workspaceConfig.User,
				},
				BuildConfig: workspaceConfig.BuildConfig,
				Repository:  repo,
				EnvVars:     workspaceConfig.EnvVars,
				PrebuildId:  prebuild.Id,
			})
		}
	}

	for _, build := range buildsToTrigger {
		createBuildDto := build_dto.BuildCreationData{
			Image:       build.ContainerConfig.Image,
			User:        build.ContainerConfig.User,
			BuildConfig: build.BuildConfig,
			Repository:  build.Repository,
			EnvVars:     build.EnvVars,
			PrebuildId:  build.PrebuildId,
		}

		_, err = s.buildService.Create(createBuildDto)
		if err != nil {
			return fmt.Errorf("failed to create build: %s", err)
		}
	}

	return nil
}

// Marks the [retention] oldest published builds for deletion for each prebuild
func (s *WorkspaceConfigService) EnforceRetentionPolicy() error {
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
				}, false)
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

func (s *WorkspaceConfigService) StartRetentionPoller() error {
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
