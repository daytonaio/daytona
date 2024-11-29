// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds_test

import (
	"testing"

	build_internal "github.com/daytonaio/daytona/internal/testing/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
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
	State: models.BuildStatePublished,
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
	State: models.BuildStatePublished,
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
	State: models.BuildStatePendingRun,
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
	State: models.BuildStatePendingRun,
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

	s.buildStore = build_internal.NewInMemoryBuildStore()
	s.buildService = builds.NewBuildService(builds.BuildServiceConfig{
		BuildStore: s.buildStore,
	})

	for _, b := range expectedBuilds {
		_ = s.buildStore.Save(b)
	}
}

func TestBuildService(t *testing.T) {
	suite.Run(t, NewBuildServiceTestSuite())
}

func (s *BuildServiceTestSuite) TestList() {
	require := s.Require()

	builds, err := s.buildService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedBuilds, builds)
}

func (s *BuildServiceTestSuite) TestFind() {
	require := s.Require()

	build, err := s.buildService.Find(&stores.BuildFilter{
		Id: &build1.Id,
	})
	require.Nil(err)
	require.Equal(build1, build)
}

func (s *BuildServiceTestSuite) TestSave() {
	expectedBuilds = append(expectedBuilds, build4)

	require := s.Require()

	// FIXME: fix me
	createBuildDto := services.CreateBuildDTO{
		WorkspaceTemplateName: "workspaceTemplateName",
		Branch:                "branch",
		PrebuildId:            &build4.PrebuildId,
		EnvVars:               build4.EnvVars,
	}

	_, err := s.buildService.Create(createBuildDto)
	require.Nil(err)

	_, err = s.buildService.List(nil)
	require.Nil(err)
	require.Contains(expectedBuilds, build4)
}

func (s *BuildServiceTestSuite) TestMarkForDeletion() {
	expectedBuilds = append(expectedBuilds, build3)

	require := s.Require()

	err := s.buildService.MarkForDeletion(&stores.BuildFilter{
		Id: &build3.Id,
	}, false)
	require.Nil(err)

	b, errs := s.buildService.Find(&stores.BuildFilter{
		Id: &build3.Id,
	})
	require.Nil(errs)
	require.Equal(b.State, models.BuildStatePendingDelete)
}

func (s *BuildServiceTestSuite) TestDelete() {
	expectedBuilds = expectedBuilds[:2]

	require := s.Require()

	err := s.buildService.Delete(build3.Id)
	require.Nil(err)

	builds, err := s.buildService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedBuilds, builds)
}
