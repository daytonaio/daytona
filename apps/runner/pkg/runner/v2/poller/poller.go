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

	apiclient "github.com/daytonaio/apiclient"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/runner/v2/executor"
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
	client      *apiclient.APIClient
}

// NewService creates a new poller service
func NewService(cfg *PollerServiceConfig) (*Service, error) {
	apiClient, err := runnerapiclient.GetApiClient()
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
	inProgressJobs, _, err := s.client.JobsAPI.ListJobs(ctx).Status(apiclient.JOBSTATUS_IN_PROGRESS).Execute()
	if err != nil {
		// Only log error
		s.log.Warn("Failed to fetch IN_PROGRESS jobs", slog.Any("error", err))
	} else {
		if inProgressJobs != nil && len(inProgressJobs.Items) > 0 {
			s.log.Info("Found IN_PROGRESS jobs", slog.Int("count", len(inProgressJobs.Items)))
			for _, job := range inProgressJobs.Items {
				go s.executor.Execute(ctx, &job)
			}
		} else {
			s.log.Info("No IN_PROGRESS jobs found")
		}
	}

	s.log.Info("Starting job poller")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Job poller stopped")
			return
		default:
			// Poll for jobs
			jobs, err := s.pollJobs(ctx)
			if err != nil {
				s.log.Warn("Failed to poll jobs", slog.Any("error", err))
				// Wait a bit before retrying on error
				time.Sleep(5 * time.Second)
				continue
			}

			// Process jobs
			if len(jobs) > 0 {
				s.log.Debug("Received jobs", slog.Int("count", len(jobs)))
				for _, job := range jobs {
					// Execute job in goroutine for parallel processing
					go s.executor.Execute(ctx, &job)
				}
			}
		}
	}
}

// pollJobs polls the API for pending jobs
func (s *Service) pollJobs(ctx context.Context) ([]apiclient.Job, error) {
	// Build poll request
	timeout := float32(s.pollTimeout.Seconds())
	limit := float32(s.pollLimit)

	req := s.client.JobsAPI.PollJobs(ctx).
		Timeout(timeout).
		Limit(limit)

	// Execute poll request
	resp, httpResp, err := req.Execute()
	if err != nil {
		// Check if it's a timeout (expected for long polling)
		if httpResp != nil && httpResp.StatusCode == 408 {
			// Timeout is normal for long polling, just return empty
			return []apiclient.Job{}, nil
		}
		return nil, err
	}

	if resp == nil {
		return []apiclient.Job{}, nil
	}

	return resp.GetJobs(), nil
}
