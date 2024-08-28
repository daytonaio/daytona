// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

type BuildRunnerInstanceConfig struct {
	Interval          string
	Scheduler         scheduler.IScheduler
	BuildRunnerId     string
	ContainerRegistry *containerregistry.ContainerRegistry
	GitProviderStore  GitProviderStore
	BuildStore        Store
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
	containerRegistry *containerregistry.ContainerRegistry
	gitProviderStore  GitProviderStore
	buildStore        Store
	builderFactory    IBuilderFactory
	loggerFactory     logs.LoggerFactory
	basePath          string
	telemetryEnabled  bool
	telemetryService  telemetry.TelemetryService
}

type BuildProcessConfig struct {
	Builder     IBuilder
	BuildLogger logs.Logger
	Build       *Build
	ProjectDir  string
	GitService  git.IGitService
	Wg          *sync.WaitGroup
}

type GitProviderStore interface {
	GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error)
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
	builds, err := r.buildStore.List(&Filter{
		States: &[]BuildState{BuildStatePendingRun, BuildStatePublished},
	})
	if err != nil {
		log.Error(err)
		return
	}

	var wg sync.WaitGroup
	for _, b := range builds {
		if b.State == BuildStatePendingRun {
			wg.Add(1)

			if b.BuildConfig == nil {
				return
			}

			buildLogger := r.loggerFactory.CreateBuildLogger(b.Id, logs.LogSourceBuilder)
			defer buildLogger.Close()

			projectDir := filepath.Join(r.basePath, b.Id, "project")

			builder, err := r.builderFactory.Create(*b, projectDir)
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
				b.State = BuildStatePublished
				err = r.buildStore.Save(b)
				if err != nil {
					r.handleBuildError(*b, builder, err, buildLogger)
					return
				}
				return
			}

			b.BuildConfig.CachedBuild = GetCachedBuild(b, builds)

			go r.RunBuildProcess(BuildProcessConfig{
				Builder:     builder,
				BuildLogger: buildLogger,
				Build:       b,
				ProjectDir:  projectDir,
				GitService: &git.Service{
					ProjectDir: projectDir,
					LogWriter:  buildLogger,
				},
				Wg: &wg,
			})
		}
	}

	wg.Wait()
}

func (r *BuildRunner) DeleteBuilds() {
	markedForDeletionBuilds, err := r.buildStore.List(&Filter{
		States: &[]BuildState{BuildStatePendingDelete},
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

		go func(b *Build) {
			buildLogger := r.loggerFactory.CreateBuildLogger(b.Id, logs.LogSourceBuilder)
			defer buildLogger.Close()

			b.State = BuildStateDeleting
			err = r.buildStore.Save(b)
			if err != nil {
				r.handleBuildError(*b, nil, err, buildLogger)
				return
			}

			err := dockerClient.DeleteImage(b.Image, true, nil)
			if err != nil {
				log.Error(err)
			}

			err = r.buildStore.Delete(b.Id)
			if err != nil {
				log.Error(err)
			}
		}(b)
	}

	wg.Wait()
}

func (r *BuildRunner) RunBuildProcess(config BuildProcessConfig) {
	if config.Wg != nil {
		defer config.Wg.Done()
	}

	config.Build.State = BuildStateRunning
	err := r.buildStore.Save(config.Build)
	if err != nil {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	gitProvider, err := r.gitProviderStore.GetConfigForUrl(config.Build.Repository.Url)
	if err != nil && !gitprovider.IsGitProviderNotFound(err) {
		r.handleBuildError(*config.Build, config.Builder, err, config.BuildLogger)
		return
	}

	var auth *http.BasicAuth
	if gitProvider != nil {
		auth = &http.BasicAuth{}
		auth.Username = gitProvider.Username
		auth.Password = gitProvider.Token
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

	config.Build.Image = image
	config.Build.User = user
	config.Build.State = BuildStateSuccess
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

	config.Build.State = BuildStatePublished
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

func (r *BuildRunner) handleBuildError(b Build, builder IBuilder, err error, buildLogger logs.Logger) {
	var errMsg string
	errMsg += "################################################\n"
	errMsg += fmt.Sprintf("#### BUILD FAILED FOR %s: %s\n", b.Id, err.Error())
	errMsg += "################################################\n"

	b.State = BuildStateError
	err = r.buildStore.Save(&b)
	if err != nil {
		errMsg += fmt.Sprintf("Error saving build: %s\n", err.Error())
	}

	cleanupErr := builder.CleanUp()
	if cleanupErr != nil {
		errMsg += fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())
	}

	buildLogger.Write([]byte(errMsg + "\n"))

	if r.telemetryEnabled {
		r.logTelemetry(context.Background(), b, err)
	}
}

func (r *BuildRunner) logTelemetry(ctx context.Context, b Build, err error) {
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
