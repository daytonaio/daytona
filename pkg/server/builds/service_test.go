// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds_test

import (
	"context"
	"testing"

	build_internal "github.com/daytonaio/daytona/internal/testing/build"
	"github.com/daytonaio/daytona/internal/testing/job"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/stretchr/testify/suite"
)

var build1Image = "image1"
var build1User = "user1"

var build1 *models.Build = &models.Build{
	Id: "id1",
	ContainerConfig: models.ContainerConfig{
		Image: build1Image,
		User:  build1User,
	},
	BuildConfig: &models.BuildConfig{},
	Repository: &gitprovider.GitRepository{
		Sha: "sha1",
	},
}

var build2 *models.Build = &models.Build{
	Id: "id2",
	ContainerConfig: models.ContainerConfig{
		Image: "image2",
		User:  "user2",
	},
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Sha: "sha2",
	},
}

var build3 *models.Build = &models.Build{
	Id: "id3",
	ContainerConfig: models.ContainerConfig{
		Image: "image3",
		User:  "user3",
	},
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Sha: "sha3",
	},
}

var build4 *models.Build = &models.Build{
	Id: "id4",
	ContainerConfig: models.ContainerConfig{
		Image: "image4",
		User:  "user4",
	},
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Sha: "sha4",
	},
}

var workspaceTemplate = &models.WorkspaceTemplate{
	Name:          "workspaceTemplateName",
	RepositoryUrl: "repositoryUrl",
	Image:         "image",
	User:          "user",
	BuildConfig:   &models.BuildConfig{},
}

var expectedBuilds []*models.Build
var expectedFilteredBuilds []*models.Build

var expectedBuildsMap map[string]*models.Build
var expectedFilteredBuildsMap map[string]*models.Build

type BuildServiceTestSuite struct {
	suite.Suite
	buildService services.IBuildService
	buildStore   stores.BuildStore
}

func NewBuildServiceTestSuite() *BuildServiceTestSuite {
	return &BuildServiceTestSuite{}
}

func (s *BuildServiceTestSuite) SetupTest() {
	expectedBuilds = []*models.Build{
		build1, build2, build3,
	}

	expectedBuildsMap = map[string]*models.Build{
		build1.Id: build1,
		build2.Id: build2,
		build3.Id: build3,
	}

	expectedFilteredBuilds = []*models.Build{
		build1, build2,
	}

	expectedFilteredBuildsMap = map[string]*models.Build{
		build1.Id: build1,
		build2.Id: build2,
	}

	jobStore := job.NewInMemoryJobStore()

	s.buildStore = build_internal.NewInMemoryBuildStore(jobStore)
	s.buildService = builds.NewBuildService(builds.BuildServiceConfig{
		BuildStore: s.buildStore,
		TrackTelemetryEvent: func(event telemetry.Event, clientId string) error {
			return nil
		},
		FindWorkspaceTemplate: func(ctx context.Context, name string) (*models.WorkspaceTemplate, error) {
			return workspaceTemplate, nil
		},
		GetRepositoryContext: func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error) {
			return &gitprovider.GitRepository{
				Url:    url,
				Branch: branch,
			}, nil
		},
		CreateJob: func(ctx context.Context, buildId string, action models.JobAction) error {
			return jobStore.Save(ctx, &models.Job{
				Id:           buildId,
				ResourceId:   buildId,
				ResourceType: models.ResourceTypeRunner,
				Action:       action,
				State:        models.JobStateSuccess,
			})
		},
	})

	for _, b := range expectedBuilds {
		_ = s.buildStore.Save(context.TODO(), b)
	}
}

func TestBuildService(t *testing.T) {
	suite.Run(t, NewBuildServiceTestSuite())
}

func (s *BuildServiceTestSuite) TestList() {
	require := s.Require()

	builds, err := s.buildService.List(context.TODO(), nil)
	require.Nil(err)
	require.Len(builds, len(expectedBuilds))
}

func (s *BuildServiceTestSuite) TestFind() {
	require := s.Require()

	build, err := s.buildService.Find(context.TODO(), &services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			Id: &build1.Id,
		},
	})
	require.Nil(err)
	require.Equal(build1.Id, build.Id)
}

func (s *BuildServiceTestSuite) TestSave() {
	expectedBuilds = append(expectedBuilds, build4)

	require := s.Require()

	createBuildDto := services.CreateBuildDTO{
		WorkspaceTemplateName: workspaceTemplate.Name,
		Branch:                "branch",
		PrebuildId:            build4.PrebuildId,
		EnvVars:               build4.EnvVars,
	}

	_, err := s.buildService.Create(context.TODO(), createBuildDto)
	require.Nil(err)

	_, err = s.buildService.List(context.TODO(), nil)
	require.Nil(err)
	require.Contains(expectedBuilds, build4)
}

func (s *BuildServiceTestSuite) TestDelete() {
	expectedBuilds = append(expectedBuilds, build3)

	require := s.Require()

	err := s.buildService.Delete(context.TODO(), &services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			Id: &build3.Id,
		},
	}, false)
	require.Nil(err)
}

func (s *BuildServiceTestSuite) TestHandleSuccessfulRemoval() {
	require := s.Require()

	err := s.buildService.HandleSuccessfulRemoval(context.TODO(), build3.Id)
	require.Nil(err)

	builds, err := s.buildService.List(context.TODO(), nil)
	require.Nil(err)
	require.NotContains(builds, build3)
}
