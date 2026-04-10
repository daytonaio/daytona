/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package poller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/daytonaio/runner/pkg/runner/v2/client"
	"github.com/daytonaio/runner/pkg/runner/v2/executor"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
)

type PollerServiceConfig struct {
	PollTimeout time.Duration
	PollLimit   int
	Logger      *slog.Logger
	Executor    *executor.Executor
}

// Service handles job polling from the API
type Service struct {
	log         *slog.Logger
	pollTimeout time.Duration
	pollLimit   int
	executor    *executor.Executor
	client      *client.APIClient
}

// NewService creates a new poller service
func NewService(cfg *PollerServiceConfig) (*Service, error) {
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Service{
		log:         cfg.Logger.With(slog.String("component", "poller")),
		pollTimeout: cfg.PollTimeout,
		pollLimit:   cfg.PollLimit,
		executor:    cfg.Executor,
		client:      apiClient,
	}, nil
}

// Start begins the job polling loop
func (s *Service) Start(ctx context.Context) {
	var listResp specsgen.ListJobsResponse
	_, err := s.client.Do(ctx, "GET", fmt.Sprintf("/jobs?status=IN_PROGRESS&limit=%d", s.pollLimit), nil, &listResp)
	if err != nil {
		s.log.WarnContext(ctx, "Failed to fetch IN_PROGRESS jobs", "error", err)
	} else {
		if len(listResp.GetItems()) > 0 {
			s.log.InfoContext(ctx, "Found IN_PROGRESS jobs", "count", len(listResp.GetItems()))
			for _, job := range listResp.GetItems() {
				go s.executor.Execute(ctx, job)
			}
		} else {
			s.log.InfoContext(ctx, "No IN_PROGRESS jobs found")
		}
	}

	s.log.InfoContext(ctx, "Starting job poller")

	for {
		select {
		case <-ctx.Done():
			s.log.InfoContext(ctx, "Job poller stopped")
			return
		default:
			jobs, err := s.pollJobs(ctx)
			if err != nil {
				s.log.ErrorContext(ctx, "Failed to poll jobs", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(jobs) > 0 {
				s.log.DebugContext(ctx, "Received jobs", "count", len(jobs))
				for _, job := range jobs {
					go s.executor.Execute(ctx, job)
				}
			}
		}
	}
}

// pollJobs polls the API for pending jobs
func (s *Service) pollJobs(ctx context.Context) ([]*specsgen.Job, error) {
	var resp specsgen.PollJobsResponse
	path := fmt.Sprintf("/jobs/poll?timeout=%d&limit=%d", int(s.pollTimeout.Seconds()), s.pollLimit)

	httpResp, err := s.client.Do(ctx, "GET", path, nil, &resp)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 408 {
			return nil, nil
		}
		return nil, err
	}

	return resp.GetJobs(), nil
}
