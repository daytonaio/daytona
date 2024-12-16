// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/jobs/build"
	"github.com/daytonaio/daytona/pkg/jobs/runner"
	"github.com/daytonaio/daytona/pkg/jobs/target"
	"github.com/daytonaio/daytona/pkg/jobs/workspace"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

// TODO: add lock when running interval func
// 1 second interval
const DEFAULT_JOB_POLL_INTERVAL = "*/1 * * * * *"

const RUNNER_METADATA_UPDATE_INTERVAL = 2 * time.Second

type IRunner interface {
	Start(ctx context.Context) error
	CheckAndRunJobs(ctx context.Context) error
}

type RunnerConfig struct {
	models.Runner
	Config    *Config
	LogWriter io.Writer

	ProviderManager providermanager.IProviderManager
	RegistryUrl     string

	ListPendingJobs   func(ctx context.Context) ([]*models.Job, int, error)
	UpdateJobState    func(ctx context.Context, jobId string, state models.JobState, err *error) error
	SetRunnerMetadata func(ctx context.Context, runnerId string, metadata models.RunnerMetadata) error

	WorkspaceJobFactory workspace.IWorkspaceJobFactory
	TargetJobFactory    target.ITargetJobFactory
	BuildJobFactory     build.IBuildJobFactory
	RunnerJobFactory    runner.IRunnerJobFactory
}

func NewRunner(config RunnerConfig) IRunner {
	return &Runner{
		Runner:    config.Runner,
		Config:    config.Config,
		LogWriter: config.LogWriter,

		providerManager: config.ProviderManager,
		registryUrl:     config.RegistryUrl,

		listPendingJobs:   config.ListPendingJobs,
		updateJobState:    config.UpdateJobState,
		setRunnerMetadata: config.SetRunnerMetadata,

		workspaceJobFactory: config.WorkspaceJobFactory,
		targetJobFactory:    config.TargetJobFactory,
		buildJobFactory:     config.BuildJobFactory,
		runnerJobFactory:    config.RunnerJobFactory,
	}
}

type Runner struct {
	models.Runner
	Config    *Config
	LogWriter io.Writer
	startTime time.Time

	providerManager providermanager.IProviderManager
	registryUrl     string

	listPendingJobs   func(ctx context.Context) ([]*models.Job, int, error)
	updateJobState    func(ctx context.Context, jobId string, state models.JobState, err *error) error
	setRunnerMetadata func(ctx context.Context, runnerId string, metadata models.RunnerMetadata) error

	workspaceJobFactory workspace.IWorkspaceJobFactory
	targetJobFactory    target.ITargetJobFactory
	buildJobFactory     build.IBuildJobFactory
	runnerJobFactory    runner.IRunnerJobFactory
}

func (r *Runner) Start(ctx context.Context) error {
	if r.Config.Id != "local" {
		r.initLogs()
		log.SetLevel(log.InfoLevel)
	}

	log.Info(fmt.Sprintf("Starting runner %s\n", r.Config.Id))

	r.startTime = time.Now()

	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		for range interruptChannel {
			plugin.CleanupClients()
		}
	}()

	// Terminate orphaned provider processes
	err := r.providerManager.TerminateProviderProcesses(r.Config.ProvidersDir)
	if err != nil {
		log.Errorf("Failed to terminate orphaned provider processes: %s", err)
	}

	err = r.downloadDefaultProviders(r.registryUrl)
	if err != nil {
		return err
	}

	err = r.registerProviders(r.registryUrl)
	if err != nil {
		return err
	}

	scheduler := scheduler.NewCronScheduler()

	err = scheduler.AddFunc(DEFAULT_JOB_POLL_INTERVAL, func() {
		err := r.CheckAndRunJobs(ctx)
		if err != nil {
			log.Error(err)
		}
	})
	if err != nil {
		return err
	}

	scheduler.Start()

	go func() {
		for {
			_ = r.UpdateRunnerMetadata(r.Config)
			time.Sleep(RUNNER_METADATA_UPDATE_INTERVAL)
		}
	}()

	log.Info("Runner started")

	<-ctx.Done()

	log.Info("Shutting down runner")
	scheduler.Stop()
	return nil
}

func (r *Runner) CheckAndRunJobs(ctx context.Context) error {
	jobs, statusCode, err := r.listPendingJobs(ctx)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	// goroutines, sync group
	for _, job := range jobs {
		err = r.runJob(ctx, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runJob(ctx context.Context, j *models.Job) error {
	var job jobs.IJob

	log.Info(fmt.Sprintf("Running job %s - %s - %s\n", j.Id, j.ResourceType, j.Action))

	err := r.updateJobState(ctx, j.Id, models.JobStateRunning, nil)
	if err != nil {
		return err
	}

	switch j.ResourceType {
	case models.ResourceTypeWorkspace:
		job = r.workspaceJobFactory.Create(*j)
	case models.ResourceTypeTarget:
		job = r.targetJobFactory.Create(*j)
	case models.ResourceTypeBuild:
		job = r.buildJobFactory.Create(*j)
	case models.ResourceTypeRunner:
		job = r.runnerJobFactory.Create(*j)
	default:
		return errors.New("invalid resource type for job")
	}

	err = job.Execute(ctx)
	if err != nil {
		log.Info(fmt.Sprintf("Job failed %s - %s - %s\n", j.Id, j.ResourceType, j.Action))
		return r.updateJobState(ctx, j.Id, models.JobStateError, &err)
	}

	log.Info(fmt.Sprintf("Job successful %s - %s - %s\n", j.Id, j.ResourceType, j.Action))
	return r.updateJobState(ctx, j.Id, models.JobStateSuccess, nil)
}

// Runner uptime in seconds
func (r *Runner) uptime() int32 {
	return max(int32(time.Since(r.startTime).Seconds()), 1)
}

func (r *Runner) UpdateRunnerMetadata(config *Config) error {
	providers := r.providerManager.GetProviders()
	uptime := r.uptime()

	providerInfos := []models.ProviderInfo{}
	for _, provider := range providers {
		info, err := provider.GetInfo()
		if err != nil {
			log.Info(fmt.Errorf("failed to get provider: %w", err))
			continue
		}

		providerInfos = append(providerInfos, info)
	}

	return r.setRunnerMetadata(context.Background(), r.Config.Id, models.RunnerMetadata{
		Uptime:      uint64(uptime),
		Providers:   providerInfos,
		RunningJobs: util.Pointer(uint64(0)),
	})
}

func (r *Runner) initLogs() {
	logFormatter := &util.LogFormatter{
		TextFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		ProcessLogWriter: r.LogWriter,
	}

	log.SetFormatter(logFormatter)
}
