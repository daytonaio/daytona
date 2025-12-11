/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package poller

import (
	"context"
	"log/slog"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-service/internal/config"
	"github.com/daytonaio/runner-service/internal/executor"
)

// Service handles job polling from the API
type Service struct {
	log      *slog.Logger
	cfg      *config.Config
	client   *apiclient.APIClient
	executor *executor.Executor
}

// NewService creates a new poller service
func NewService(cfg *config.Config, client *apiclient.APIClient, executor *executor.Executor, logger *slog.Logger) *Service {
	return &Service{
		log:      logger.With(slog.String("component", "poller")),
		cfg:      cfg,
		client:   client,
		executor: executor,
	}
}

// Start begins the job polling loop
func (s *Service) Start(ctx context.Context) {
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
				s.log.Info("Received jobs", slog.Int("count", len(jobs)))
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
	timeout := float32(s.cfg.PollTimeout.Seconds())
	limit := float32(s.cfg.PollLimit)

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
