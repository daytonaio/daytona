// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jobs_test

import (
	"testing"

	job_internal "github.com/daytonaio/daytona/internal/testing/job"
	"github.com/daytonaio/daytona/pkg/models"
	jobs "github.com/daytonaio/daytona/pkg/server/jobs"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

var expectedJobs []*models.Job
var expectedFilteredJobs []*models.Job

var expectedJobsMap map[string]*models.Job
var expectedFilteredJobsMap map[string]*models.Job

var job1 = &models.Job{
	Id:         "1",
	ResourceId: "1",
	Action:     models.JobActionStart,
	State:      models.JobStatePending,
}

var job2 = &models.Job{
	Id:         "2",
	ResourceId: "2",
	Action:     models.JobActionStart,
	State:      models.JobStatePending,
}

var job3 = &models.Job{
	Id:         "3",
	ResourceId: "3",
	Action:     models.JobActionStart,
	State:      models.JobStatePending,
}

var job4 = &models.Job{
	Id:         "4",
	ResourceId: "4",
	Action:     models.JobActionStart,
	State:      models.JobStatePending,
}

type JobServiceTestSuite struct {
	suite.Suite
	jobService services.IJobService
	jobStore   stores.JobStore
}

func NewJobServiceTestSuite() *JobServiceTestSuite {
	return &JobServiceTestSuite{}
}

func (s *JobServiceTestSuite) SetupTest() {
	expectedJobs = []*models.Job{
		job1, job2, job3,
	}

	expectedJobsMap = map[string]*models.Job{
		job1.Id: job1,
		job2.Id: job2,
		job3.Id: job3,
	}

	expectedFilteredJobs = []*models.Job{
		job1, job2,
	}

	expectedFilteredJobsMap = map[string]*models.Job{
		job1.Id: job1,
		job2.Id: job2,
	}

	s.jobStore = job_internal.NewInMemoryJobStore()
	s.jobService = jobs.NewJobService(jobs.JobServiceConfig{
		JobStore: s.jobStore,
	})

	for _, j := range expectedJobs {
		_ = s.jobStore.Save(j)
	}
}

func TestJobService(t *testing.T) {
	suite.Run(t, NewJobServiceTestSuite())
}

func (s *JobServiceTestSuite) TestList() {
	require := s.Require()

	jobs, err := s.jobService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}

func (s *JobServiceTestSuite) TestFind() {
	require := s.Require()

	job, err := s.jobService.Find(&stores.JobFilter{
		Id: &job1.Id,
	})
	require.Nil(err)
	require.Equal(job1, job)
}

func (s *JobServiceTestSuite) TestSave() {
	expectedJobs = append(expectedJobs, job4)

	require := s.Require()

	err := s.jobService.Save(job4)
	require.Nil(err)

	jobs, err := s.jobService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}

func (s *JobServiceTestSuite) TestDelete() {
	expectedJobs = expectedJobs[:2]

	require := s.Require()

	err := s.jobService.Delete(job3)
	require.Nil(err)

	jobs, err := s.jobService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}
