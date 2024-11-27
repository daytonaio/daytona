//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package job

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryJobStore struct {
	jobs map[string]*models.Job
}

func NewInMemoryJobStore() stores.JobStore {
	return &InMemoryJobStore{
		jobs: make(map[string]*models.Job),
	}
}

func (s *InMemoryJobStore) List(filter *stores.JobFilter) ([]*models.Job, error) {
	jobs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *InMemoryJobStore) Find(filter *stores.JobFilter) (*models.Job, error) {
	jobs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, stores.ErrJobNotFound
	}

	return jobs[0], nil
}

func (s *InMemoryJobStore) Save(job *models.Job) error {
	s.jobs[job.Id] = job
	return nil
}

func (s *InMemoryJobStore) Delete(job *models.Job) error {
	delete(s.jobs, job.Id)
	return nil
}

func (s *InMemoryJobStore) processFilters(filter *stores.JobFilter) ([]*models.Job, error) {
	var result []*models.Job
	filteredJobs := make(map[string]*models.Job)
	for k, v := range s.jobs {
		filteredJobs[k] = v
	}

	if filter != nil {
		if filter.Id != nil {
			job, ok := s.jobs[*filter.Id]
			if ok {
				return []*models.Job{job}, nil
			} else {
				return []*models.Job{}, fmt.Errorf("job with id %s not found", *filter.Id)
			}
		}
		if filter.States != nil {
			for _, job := range filteredJobs {
				check := false
				for _, state := range *filter.States {
					if job.State == state {
						check = true
						break
					}
				}
				if !check {
					delete(filteredJobs, job.Id)
				}
			}
		}
		if filter.Actions != nil {
			for _, job := range filteredJobs {
				check := false
				for _, action := range *filter.Actions {
					if job.Action == action {
						check = true
						break
					}
				}
				if !check {
					delete(filteredJobs, job.Id)
				}
			}
		}
	}

	for _, job := range filteredJobs {
		result = append(result, job)
	}

	return result, nil
}
