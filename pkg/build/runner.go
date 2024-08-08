// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"sync"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

type BuildRunnerInstanceConfig struct {
	Interval         string
	Scheduler        scheduler.IScheduler
	BuildRunnerId    string
	BuildStore       Store
	BuilderFactory   IBuilderFactory
	LoggerFactory    logs.LoggerFactory
	TelemetryEnabled bool
	TelemetryService telemetry.TelemetryService
}

type BuildRunner struct {
	Id               string
	scheduler        scheduler.IScheduler
	interval         string
	buildStore       Store
	builderFactory   IBuilderFactory
	loggerFactory    logs.LoggerFactory
	telemetryEnabled bool
	telemetryService telemetry.TelemetryService
}

func NewBuildRunner(config BuildRunnerInstanceConfig) *BuildRunner {
	runner := &BuildRunner{
		Id:               config.BuildRunnerId,
		scheduler:        config.Scheduler,
		interval:         config.Interval,
		buildStore:       config.BuildStore,
		builderFactory:   config.BuilderFactory,
		loggerFactory:    config.LoggerFactory,
		telemetryEnabled: config.TelemetryEnabled,
		telemetryService: config.TelemetryService,
	}

	return runner
}

func (r *BuildRunner) Start() error {
	err := r.scheduler.AddFunc(r.interval, func() { r.Run() })
	if err != nil {
		return err
	}

	r.scheduler.Start()

	return nil
}

func (r *BuildRunner) Stop() {
	r.scheduler.Stop()
}

func (r *BuildRunner) Run() {
	pendingState := BuildStatePending
	builds, err := r.buildStore.List(&BuildFilter{State: &pendingState})
	if err != nil {
		log.Error(err)
		return
	}

	var wg sync.WaitGroup
	for _, build := range builds {
		wg.Add(1)
		go r.RunBuildProcess(build, &wg)
	}

	wg.Wait()

}

func (r *BuildRunner) RunBuildProcess(build *Build, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	if build.Project.BuildConfig == nil {
		return
	}

	buildLogger := r.loggerFactory.CreateBuildLogger(build.Project.Name, build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	builder, err := r.builderFactory.Create(*build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	build.State = BuildStateRunning
	err = r.buildStore.Save(build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	image, user, err := builder.Build(*build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	build.Image = image
	build.User = user
	build.State = BuildStateSuccess
	err = r.buildStore.Save(build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	err = builder.Publish(*build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	build.State = BuildStatePublished
	err = r.buildStore.Save(build)
	if err != nil {
		r.handleBuildError(*build, builder, err, buildLogger)
		return
	}

	err = builder.CleanUp()
	if err != nil {
		errMsg := fmt.Sprintf("Error cleaning up build: %s\n", err.Error())
		buildLogger.Write([]byte(errMsg + "\n"))
	}

	if r.telemetryEnabled {
		r.logTelemetry(context.Background(), *build, err)
	}

}

func (r *BuildRunner) handleBuildError(build Build, builder IBuilder, err error, buildLogger logs.Logger) {
	var errMsg string
	errMsg += "################################################\n"
	errMsg += fmt.Sprintf("#### BUILD FAILED FOR PROJECT %s: %s\n", build.Project.Name, err.Error())
	errMsg += "################################################\n"

	build.State = BuildStateError
	err = r.buildStore.Save(&build)
	if err != nil {
		errMsg += fmt.Sprintf("Error saving build: %s\n", err.Error())
	}

	cleanupErr := builder.CleanUp()
	if cleanupErr != nil {
		errMsg += fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())
	}

	buildLogger.Write([]byte(errMsg + "\n"))

	if r.telemetryEnabled {
		r.logTelemetry(context.Background(), build, err)
	}
}

func (r *BuildRunner) logTelemetry(ctx context.Context, build Build, err error) {
	telemetryProps := telemetry.NewBuildRunnerEventProps(ctx, build.Id, string(build.State))
	event := telemetry.BuildEventRunBuild
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.BuildEventRunBuildError
	}
	telemetryError := r.telemetryService.TrackBuildEvent(event, r.Id, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}
}
