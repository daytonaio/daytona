// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (bj *BuildJob) run(ctx context.Context, j *models.Job) error {
	b, err := bj.findBuild(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	successfulBuilds, err := bj.listSuccessfulBuilds(ctx, b.Repository.Url)
	if err != nil {
		return err
	}

	if b.BuildConfig == nil {
		return fmt.Errorf("build config is not set")
	}

	buildLogger, err := bj.loggerFactory.CreateBuildLogger(b.Id, logs.LogSourceBuilder)
	if err != nil {
		return err
	}
	defer buildLogger.Close()

	workspaceDir := filepath.Join(bj.basePath, b.Id, "workspace")

	builder, err := bj.builderFactory.Create(b.Build, workspaceDir)
	if err != nil {
		return bj.handleBuildResult(b.Build, builder, buildLogger, err)
	}

	imageName, err := builder.GetImageName(b.Build)
	if err != nil {
		return bj.handleBuildResult(b.Build, builder, buildLogger, err)
	}

	exists := bj.checkImageExists(ctx, imageName)
	if err != nil {
		return bj.handleBuildResult(b.Build, builder, buildLogger, err)
	}

	if exists {
		return bj.handleBuildResult(b.Build, builder, buildLogger, nil)
	}

	b.BuildConfig.CachedBuild = models.GetCachedBuild(&b.Build, successfulBuilds)

	err = bj.runBuildProcess(ctx, BuildProcessConfig{
		Builder: builder,
		Build:   &b.Build,
		GitService: &git.Service{
			WorkspaceDir: workspaceDir,
			LogWriter:    buildLogger,
		},
	})
	if err != nil {
		return bj.handleBuildResult(b.Build, builder, buildLogger, err)
	}

	buildLogger.Write([]byte("\n \n" + lipgloss.NewStyle().Bold(true).Render("Build run successful. Publishing...")))

	err = builder.Publish(b.Build)
	if err != nil {
		return bj.handleBuildResult(b.Build, builder, buildLogger, err)
	}

	err = builder.CleanUp()
	if err != nil {
		errMsg := fmt.Sprintf("Error cleaning up build: %s\n", err.Error())
		buildLogger.Write([]byte(errMsg + "\n"))
	}

	return bj.handleBuildResult(b.Build, builder, buildLogger, err)
}

type BuildProcessConfig struct {
	Builder    build.IBuilder
	Build      *models.Build
	GitService git.IGitService
}

func (bj *BuildJob) runBuildProcess(ctx context.Context, config BuildProcessConfig) error {
	gitProviders, err := bj.listConfigsForUrl(ctx, config.Build.Repository.Url)
	if err != nil {
		return err
	}

	var auth *http.BasicAuth
	if len(gitProviders) > 0 {
		auth = &http.BasicAuth{}
		auth.Username = gitProviders[0].Username
		auth.Password = gitProviders[0].Token
	}

	err = config.GitService.CloneRepository(config.Build.Repository, auth)
	if err != nil {
		return err
	}

	image, user, err := config.Builder.Build(*config.Build)
	if err != nil {
		return err
	}

	config.Build.Image = &image
	config.Build.User = &user

	return nil
}

func (bj *BuildJob) handleBuildResult(b models.Build, builder build.IBuilder, buildLogger logs.Logger, err error) error {
	var errMsg string

	if err != nil {
		errMsg += "################################################\n"
		errMsg += fmt.Sprintf("#### BUILD FAILED FOR %s: %s\n", b.Id, err.Error())
		errMsg += "################################################\n"
	}

	if builder != nil {
		cleanupErr := builder.CleanUp()
		if cleanupErr != nil {
			errMsg += fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())
		}
	}

	buildLogger.Write([]byte(errMsg + "\n"))
	return err
}
