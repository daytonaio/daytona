// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jobs_test

import (
	"context"
	"testing"

	job_internal "github.com/daytonaio/daytona/internal/testing/job"
	"github.com/daytonaio/daytona/pkg/models"
	jobs "github.com/daytonaio/daytona/pkg/server/jobs"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/stretchr/testify/suite"
)

var expectedJobs []*models.Job

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

	s.jobStore = job_internal.NewInMemoryJobStore()
	s.jobService = jobs.NewJobService(jobs.JobServiceConfig{
		JobStore: s.jobStore,
		TrackTelemetryEvent: func(event telemetry.Event, clientId string) error {
			return nil
		},
	})

	for _, j := range expectedJobs {
		_ = s.jobStore.Save(context.TODO(), j)
	}
}

func TestJobService(t *testing.T) {
	suite.Run(t, NewJobServiceTestSuite())
}

func (s *JobServiceTestSuite) TestList() {
	require := s.Require()

	jobs, err := s.jobService.List(context.TODO(), nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}

func (s *JobServiceTestSuite) TestFind() {
	require := s.Require()

	job, err := s.jobService.Find(context.TODO(), &stores.JobFilter{
		Id: &job1.Id,
	})
	require.Nil(err)
	require.Equal(job1, job)
}

func (s *JobServiceTestSuite) TestCreate() {
	expectedJobs = append(expectedJobs, job4)

	require := s.Require()

	err := s.jobService.Create(context.TODO(), job4)
	require.Nil(err)

	jobs, err := s.jobService.List(context.TODO(), nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}

func (s *JobServiceTestSuite) TestSetState() {
	require := s.Require()

	err := s.jobService.Create(context.TODO(), job4)
	require.Nil(err)

	job4Update := *job4
	job4Update.State = models.JobStateSuccess

	err = s.jobService.SetState(context.TODO(), job4Update.Id, services.UpdateJobStateDTO{
		State:        models.JobStateSuccess,
		ErrorMessage: nil,
	})
	require.Nil(err)

	updated, err := s.jobService.Find(context.TODO(), &stores.JobFilter{
		Id: &job4.Id,
	})
	require.Nil(err)
	require.Equal(job4Update, *updated)
}

func (s *JobServiceTestSuite) TestCreateWithAnotherJobInProgress() {
	require := s.Require()

	err := s.jobService.Create(context.TODO(), job4)
	require.Nil(err)

	var job5 = &models.Job{
		Id:         "5",
		ResourceId: "4",
		Action:     models.JobActionStart,
		State:      models.JobStatePending,
	}

	err = s.jobService.Create(context.TODO(), job5)
	require.EqualError(err, stores.ErrJobInProgress.Error())
}

func (s *JobServiceTestSuite) TestDelete() {
	expectedJobs = expectedJobs[:2]

	require := s.Require()

	err := s.jobService.Delete(context.TODO(), job3)
	require.Nil(err)

	jobs, err := s.jobService.List(context.TODO(), nil)
	require.Nil(err)
	require.ElementsMatch(expectedJobs, jobs)
}
