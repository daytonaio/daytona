// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

type BuildRunnerInstanceConfig struct {
	Interval          string
	Scheduler         scheduler.IScheduler
	BuildRunnerId     string
	ContainerRegistry *models.ContainerRegistry
	GitProviderStore  GitProviderStore
	BuildStore        builds.BuildStore
	BuilderFactory    IBuilderFactory
	LoggerFactory     logs.LoggerFactory
	BasePath          string
	TelemetryEnabled  bool
	TelemetryService  telemetry.TelemetryService
}

type BuildRunner struct {
	Id                string
	scheduler         scheduler.IScheduler
	runInterval       string
	containerRegistry *models.ContainerRegistry
	gitProviderStore  GitProviderStore
	buildStore        builds.BuildStore
	builderFactory    IBuilderFactory
	loggerFactory     logs.LoggerFactory
	basePath          string
	telemetryEnabled  bool
	telemetryService  telemetry.TelemetryService
}

type BuildProcessConfig struct {
	Builder      IBuilder
	BuildLogger  logs.Logger
	Build        *models.Build
	WorkspaceDir string
	GitService   git.IGitService
	Wg           *sync.WaitGroup
}

type GitProviderStore interface {
	ListConfigsForUrl(url string) ([]*models.GitProviderConfig, error)
}

func NewBuildRunner(config BuildRunnerInstanceConfig) *BuildRunner {
	runner := &BuildRunner{
		Id:                config.BuildRunnerId,
		scheduler:         config.Scheduler,
		runInterval:       config.Interval,
		containerRegistry: config.ContainerRegistry,
		gitProviderStore:  config.GitProviderStore,
		buildStore:        config.BuildStore,
		builderFactory:    config.BuilderFactory,
		loggerFactory:     config.LoggerFactory,
		basePath:          config.BasePath,
		telemetryEnabled:  config.TelemetryEnabled,
		telemetryService:  config.TelemetryService,
	}

	return runner
}

func (r *BuildRunner) Start() error {
	err := r.scheduler.AddFunc(r.runInterval, func() { r.RunBuilds() })
	if err != nil {
		return err
	}
	err = r.scheduler.AddFunc(r.runInterval, func() { r.DeleteBuilds() })
	if err != nil {
		return err
	}

	r.scheduler.Start()
	return nil
}

func (r *BuildRunner) Stop() {
	r.scheduler.Stop()
}

func (r *BuildRunner) RunBuilds() {
	builds, err := r.buildStore.List(&builds.BuildFilter{
		States: &[]models.BuildState{models.BuildStatePendingRun, models.BuildStatePublished},
	})
	if err != nil {
		log.Error(err)
		return
	}

	var wg sync.WaitGroup
	for _, b := range builds {
		if b.State == models.BuildStatePendingRun {
			wg.Add(1)

			if b.BuildConfig == nil {
				return
			}

			buildLogger := r.loggerFactory.CreateBuildLogger(b.Id, logs.LogSourceBuilder)
			defer buildLogger.Close()

			workspaceDir := filepath.Join(r.basePath, b.Id, "workspace")

			builder, err := r.builderFactory.Create(*b, workspaceDir)
			if err != nil {
				r.handleBuildError(*b, builder, err, buildLogger)
				return
			}

			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				log.Error(err)
				return
			}

			imageName, err := builder.GetImageName(*b)
			if err != nil {
				r.handleBuildError(*b, builder, err, buildLogger)
				return
			}

			_, _, err = cli.ImageInspectWithRaw(context.Background(), imageName)
			if err == nil {
				b.State = models.BuildStatePublished
				err = r.buildStore.Save(b)
				if err != nil {
					r.handleBuildError(*b, builder, err, buildLogger)
					return
				}
				return
			}

			b.BuildConfig.CachedBuild = models.GetCachedBuild(b, builds)

			go r.RunBuildProcess(BuildProcessConfig{
				Builder:      builder,
				BuildLogger:  buildLogger,
				Build:        b,
				WorkspaceDir: workspaceDir,
				GitService: &git.Service{
					WorkspaceDir: workspaceDir,
					LogWriter:    buildLogger,
				},
				Wg: &wg,
			})
		}
	}

	wg.Wait()
}

func (r *BuildRunner) DeleteBuilds() {
	markedForDeletionBuilds, err := r.buildStore.List(&builds.BuildFilter{
		States: &[]models.BuildState{models.BuildStatePendingDelete, models.BuildStatePendingForcedDelete},
	})
	if err != nil {
		log.Error(err)
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error(err)
		return
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	var wg sync.WaitGroup
	for _, b := range markedForDeletionBuilds {
		wg.Add(1)

		go func(b *models.Build) {
			buildLogger := r.loggerFactory.CreateBuildLogger(b.Id, logs.LogSourceBuilder)
			defer buildLogger.Close()

			force := b.State == models.BuildStatePendingForcedDelete

			b.State = models.BuildStateDeleting
			err = r.buildStore.Save(b)
			if err != nil {
				r.handleBuildError(*b, nil, err, buildLogger)
				return
			}

			// If the build has an image, delete it first
			if b.Image != nil {
				err := dockerClient.DeleteImage(*b.Image, true, nil)
				if err != nil {
					r.handleBuildError(*b, nil, err, buildLogger)
					if !force {
						return
					}
				}
			}

			err = r.buildStore.Delete(b.Id)
			if err != nil {
				r.handleBuildError(*b, nil, err, buildLogger)
				return
			}
		}(b)
	}

	wg.Wait()
}

func (r *BuildRunner) RunBuildProcess(config BuildProcessConfig) {
	if config.Wg != nil {
		defer config.Wg.Done()
	}

	config.Build.State = models.BuildStateRunning
	err := r.buildStore.Save(config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	gitProviders, err := r.gitProviderStore.ListConfigsForUrl(config.Build.Repository.Url)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	var auth *http.BasicAuth
	if len(gitProviders) > 0 {
		auth = &http.BasicAuth{}
		auth.Username = gitProviders[0].Username
		auth.Password = gitProviders[0].Token
	}

	err = config.GitService.CloneRepository(config.Build.Repository, auth)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	image, user, err := config.Builder.Build(*config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	config.Build.Image = &image
	config.Build.User = &user
	config.Build.State = models.BuildStateSuccess
	err = r.buildStore.Save(config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	err = config.Builder.Publish(*config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	config.Build.State = models.BuildStatePublished
	err = r.buildStore.Save(config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	err = config.Builder.CleanUp()
	if err != nil {
		errMsg := fmt.Sprintf("Error cleaning up build: %s\n", err.Error())
		config.BuildLogger.Write([]byte(errMsg + "\n"))
	}

	config.BuildLogger.Write([]byte("\n \n" + lipgloss.NewStyle().Bold(true).Render("Build completed successfully")))

	if r.telemetryEnabled {
		r.logTelemetry(context.Background(), *config.Build, err)
	}
}

func (r *BuildRunner) handleBuildError(b models.Build, builder IBuilder, err error, buildLogger logs.Logger) {
	var errMsg string
	errMsg += "################################################\n"
	errMsg += fmt.Sprintf("#### BUILD FAILED FOR %s: %s\n", b.Id, err.Error())
	errMsg += "################################################\n"

	b.State = models.BuildStateError
	err = r.buildStore.Save(&b)
	if err != nil {
		errMsg += fmt.Sprintf("Error saving build: %s\n", err.Error())
	}

	if builder != nil {
		cleanupErr := builder.CleanUp()
		if cleanupErr != nil {
			errMsg += fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())
		}
	}

	buildLogger.Write([]byte(errMsg + "\n"))

	if r.telemetryEnabled {
		r.logTelemetry(context.Background(), b, err)
	}
}

func (r *BuildRunner) logTelemetry(ctx context.Context, b models.Build, err error) {
	telemetryProps := telemetry.NewBuildRunnerEventProps(ctx, b.Id, string(b.State))
	event := telemetry.BuildRunnerEventRunBuild
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.BuildRunnerEventRunBuildError
	}
	telemetryError := r.telemetryService.TrackBuildRunnerEvent(event, r.Id, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}
}
