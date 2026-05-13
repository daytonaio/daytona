// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	cloneJobStatusRunning = "running"
	cloneJobStatusDone    = "done"
	cloneJobStatusError   = "error"
)

type cloneJob struct {
	ID         string
	Status     string
	Path       string
	StartedAt  time.Time
	FinishedAt *time.Time
	Error      string
}

type cloneJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*cloneJob
}

func newCloneJobStore() *cloneJobStore {
	return &cloneJobStore{jobs: map[string]*cloneJob{}}
}

func (s *cloneJobStore) create(path string) *cloneJob {
	job := &cloneJob{
		ID:        uuid.NewString(),
		Status:    cloneJobStatusRunning,
		Path:      path,
		StartedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	return job
}

func (s *cloneJobStore) finish(id string, err error) {
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return
	}

	job.FinishedAt = &now
	if err != nil {
		job.Status = cloneJobStatusError
		job.Error = err.Error()
		return
	}
	job.Status = cloneJobStatusDone
}

func (s *cloneJobStore) get(id string) (*cloneJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	if !ok {
		return nil, false
	}

	copied := *job
	return &copied, true
}

func (j *cloneJob) response() GitCloneJobResponse {
	resp := GitCloneJobResponse{
		JobID:     j.ID,
		Status:    j.Status,
		Path:      j.Path,
		StartedAt: j.StartedAt.Format(time.RFC3339Nano),
		Error:     j.Error,
	}
	if j.FinishedAt != nil {
		resp.FinishedAt = j.FinishedAt.Format(time.RFC3339Nano)
	}
	return resp
}

var cloneJobs = newCloneJobStore()

var errCloneJobNotFound = errors.New("clone expansion job not found")
