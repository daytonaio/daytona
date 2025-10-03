// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	log "github.com/sirupsen/logrus"
)

type GarbageCollectorServiceConfig struct {
	ApiBaseUrl         string
	ApiToken           string
	Domain             string
	DryRun             bool
	ExcludeSandboxes   string
	ExcludeSnapshots   string
	ThresholdSandboxes int
	ThresholdSnapshots int
	ExcludeAge         string
	Interval           string
	DockerClient       *docker.DockerClient
}

type GarbageCollectorService struct {
	apiBaseUrl         string
	apiToken           string
	domain             string
	dryRun             bool
	excludeSandboxes   []string
	excludeSnapshots   []string
	thresholdSandboxes int
	thresholdSnapshots int
	excludeAge         string
	interval           string
	dockerClient       *docker.DockerClient
	httpClient         *http.Client
	client             *apiclient.APIClient
}

func NewGarbageCollectorService(config GarbageCollectorServiceConfig) *GarbageCollectorService {
	excludeSandboxes := []string{}
	excludeSnapshots := []string{}

	excludeSandboxesRaw := strings.Split(config.ExcludeSandboxes, ",")
	for _, exclude := range excludeSandboxesRaw {
		exclude = strings.TrimSpace(exclude)
		if exclude != "" {
			excludeSandboxes = append(excludeSandboxes, exclude)
		}
	}

	excludeSnapshotsRaw := strings.Split(config.ExcludeSnapshots, ",")
	for _, exclude := range excludeSnapshotsRaw {
		exclude = strings.TrimSpace(exclude)
		if exclude != "" {
			excludeSnapshots = append(excludeSnapshots, exclude)
		}
	}

	client, err := runnerapiclient.GetApiClient()
	if err != nil {
		log.Errorf("Failed to get API client when initializing garbage collector on runner %s: %v", config.Domain, err)
	}

	return &GarbageCollectorService{
		apiBaseUrl:         config.ApiBaseUrl,
		apiToken:           config.ApiToken,
		domain:             config.Domain,
		dryRun:             config.DryRun,
		excludeSandboxes:   excludeSandboxes,
		excludeSnapshots:   excludeSnapshots,
		thresholdSandboxes: config.ThresholdSandboxes,
		thresholdSnapshots: config.ThresholdSnapshots,
		excludeAge:         config.ExcludeAge,
		interval:           config.Interval,
		dockerClient:       config.DockerClient,
		httpClient:         &http.Client{},
		client:             client,
	}
}

func (s *GarbageCollectorService) Run(ctx context.Context) {
	go func() {
		// Run cleanup immediately on startup
		_ = s.cleanup(ctx)

		// Parse interval, default to 12 hours if not available or invalid format
		tickerInterval, err := time.ParseDuration(s.interval)
		if err != nil {
			log.Errorf("Failed to parse interval: %v, using default interval of 12 hours", err)
			tickerInterval = 12 * time.Hour
		}

		// Set up ticker for every interval
		ticker := time.NewTicker(tickerInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = s.cleanup(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *GarbageCollectorService) cleanup(ctx context.Context) error {
	log.Printf("Starting sandbox and snapshot cleanup process")
	log.Printf("Dry run mode: %t", s.dryRun)
	log.Printf("Exclude pattern: %s", s.excludeSandboxes)
	log.Printf("Exclude snapshots pattern: %s", s.excludeSnapshots)
	log.Printf("Sandbox removal threshold: %d", s.thresholdSandboxes)
	log.Printf("Snapshot removal threshold: %d", s.thresholdSnapshots)

	if s.client == nil {
		client, err := runnerapiclient.GetApiClient()
		if err != nil {
			return fmt.Errorf("cleanup terminated: failed to initialize API client for garbage collector on runner %s: %w", s.domain, err)
		}
		s.client = client
	}

	// Get all sandboxes
	sandboxes, err := s.getSandboxes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sandboxes: %w", err)
	}
	log.Printf("Found %d sandboxes", len(sandboxes))

	// Filter out excluded sandboxes
	filteredSandboxes := s.filterSandboxes(sandboxes)
	log.Printf("After filtering exclusions: %d sandboxes", len(filteredSandboxes))

	// Check each sandbox against API
	invalidSandboxes := []models.SandboxCleanupInfo{}
	for _, sandbox := range filteredSandboxes {
		exists, err := s.checkSandboxInAPI(ctx, sandbox.Name)
		if err != nil {
			log.Printf("Error checking sandbox %s: %v", sandbox.Name, err)
			continue
		}
		if !exists {
			invalidSandboxes = append(invalidSandboxes, sandbox)
		}
	}

	log.Printf("Found %d invalid sandboxes", len(invalidSandboxes))

	// Get all snapshots
	snapshots, err := s.getSnapshots(ctx)
	if err != nil {
		return fmt.Errorf("failed to get snapshots: %w", err)
	}
	log.Printf("Found %d snapshots", len(snapshots))

	// Filter out excluded snapshots
	filteredSnapshots := s.filterSnapshots(snapshots)
	log.Printf("After filtering snapshot exclusions: %d snapshots", len(filteredSnapshots))

	// Check each snapshot against API
	invalidSnapshots := []models.SnapshotCleanupInfo{}
	for _, snapshot := range filteredSnapshots {
		exists, err := s.checkSnapshotInAPI(ctx, snapshot.Name)
		if err != nil {
			log.Printf("Error checking snapshot %s: %v", snapshot.Name, err)
			continue
		}
		if !exists {
			invalidSnapshots = append(invalidSnapshots, snapshot)
		}
	}

	log.Printf("Found %d invalid snapshots", len(invalidSnapshots))

	// Display results
	if len(invalidSandboxes) == 0 && len(invalidSnapshots) == 0 {
		log.Printf("No invalid sandboxes or snapshots found")
		return nil
	}

	if len(invalidSandboxes) > 0 {
		log.Printf("Invalid sandboxes:")
		for _, sandbox := range invalidSandboxes {
			log.Printf("  - %s (%s)", sandbox.Name, sandbox.ID)
		}
	}

	if len(invalidSnapshots) > 0 {
		log.Printf("Invalid snapshots:")
		for _, snapshot := range invalidSnapshots {
			log.Printf("  - %s (%s)", snapshot.Name, snapshot.ID)
		}
	}

	// Check thresholds before proceeding with deletion
	if s.thresholdSandboxes > 0 && len(invalidSandboxes) > s.thresholdSandboxes {
		return fmt.Errorf("aborting: found %d sandboxes to remove, which exceeds the sandbox threshold of %d",
			len(invalidSandboxes), s.thresholdSandboxes)
	}

	if s.thresholdSnapshots > 0 && len(invalidSnapshots) > s.thresholdSnapshots {
		return fmt.Errorf("aborting: found %d snapshots to remove, which exceeds the snapshot threshold of %d",
			len(invalidSnapshots), s.thresholdSnapshots)
	}

	if s.dryRun {
		log.Printf("Dry run mode - no sandboxes or snapshots will be removed")
		return nil
	}

	// Remove invalid sandboxes
	if len(invalidSandboxes) > 0 {
		log.Printf("Removing %d invalid sandboxes", len(invalidSandboxes))
		removedCount := 0
		for _, sandbox := range invalidSandboxes {
			// Remove the sandbox (this will stop it first if needed)
			err := s.dockerClient.ApiClient().ContainerRemove(context.Background(), sandbox.ID, container.RemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
			if err != nil {
				log.Printf("Error removing sandbox %s: %v", sandbox.Name, err)
			} else {
				log.Printf("Removed sandbox %s (%s)", sandbox.Name, sandbox.ID)
				removedCount++
			}
		}
		log.Printf("Successfully removed %d out of %d invalid sandboxes", removedCount, len(invalidSandboxes))
	}

	// Remove invalid snapshots
	if len(invalidSnapshots) > 0 {
		log.Printf("Removing %d invalid snapshots", len(invalidSnapshots))
		removedCount := 0
		for _, snapshot := range invalidSnapshots {
			if err := s.dockerClient.RemoveImage(context.Background(), snapshot.Name, true); err != nil {
				log.Printf("Error removing snapshot %s: %v", snapshot.Name, err)
			} else {
				log.Printf("Removed snapshot %s (%s)", snapshot.Name, snapshot.ID)
				removedCount++
			}
		}
		log.Printf("Successfully removed %d out of %d invalid snapshots", removedCount, len(invalidSnapshots))
	}

	return nil
}

func (s *GarbageCollectorService) checkSandboxInAPI(ctx context.Context, sandboxName string) (bool, error) {
	runnerBySandboxResponse, resp, err := s.client.RunnersAPI.GetRunnerBySandboxId(ctx, sandboxName).Execute()
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		// If 404, sandbox doesn't exist in API
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		// For other status codes, log the error and assume sandbox exist
		log.Printf("Unexpected status code %d for sandbox %s", resp.StatusCode, sandboxName)
		return false, fmt.Errorf("unexpected status code %d for sandbox %s", resp.StatusCode, sandboxName)
	}

	// Check if runner exists and domain matches
	if runnerBySandboxResponse.Id != "" && runnerBySandboxResponse.Domain == s.domain {
		return true, nil
	}

	// Runner exists but domain doesn't match
	log.Printf("Sandbox %s belongs to runner domain %s, expected %s", sandboxName, runnerBySandboxResponse.Domain, s.domain)

	return true, nil
}

func (s *GarbageCollectorService) getSandboxes(ctx context.Context) ([]models.SandboxCleanupInfo, error) {
	sandboxes, err := s.dockerClient.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var sandboxCleanupInfos []models.SandboxCleanupInfo
	for _, sandbox := range sandboxes {
		name := strings.TrimPrefix(sandbox.Names[0], "/")
		sandboxCleanupInfos = append(sandboxCleanupInfos, models.SandboxCleanupInfo{
			ID:   sandbox.ID,
			Name: name,
		})
	}

	return sandboxCleanupInfos, nil
}

func (s *GarbageCollectorService) filterSandboxes(sandboxes []models.SandboxCleanupInfo) []models.SandboxCleanupInfo {
	if len(s.excludeSandboxes) == 0 {
		return sandboxes
	}

	var filtered []models.SandboxCleanupInfo
	for _, sandbox := range sandboxes {
		if !slices.Contains(s.excludeSandboxes, strings.TrimSpace(sandbox.Name)) {
			filtered = append(filtered, sandbox)
		}
	}

	return filtered
}

func (s *GarbageCollectorService) getSnapshots(ctx context.Context) ([]models.SnapshotCleanupInfo, error) {
	snapshots, err := s.dockerClient.ApiClient().ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var snapshotCleanupInfos []models.SnapshotCleanupInfo
	for _, snapshot := range snapshots {
		var name string
		if len(snapshot.RepoTags) > 0 {
			name = snapshot.RepoTags[0]
		} else if len(snapshot.RepoDigests) > 0 {
			name = snapshot.RepoDigests[0]
		} else {
			// Use short ID if no tags or digests
			name = snapshot.ID
			if len(name) > 12 {
				name = name[:12]
			}
		}

		snapshotCleanupInfos = append(snapshotCleanupInfos, models.SnapshotCleanupInfo{
			ID:        snapshot.ID,
			Name:      name,
			CreatedAt: time.Unix(snapshot.Created, 0),
		})
	}

	return snapshotCleanupInfos, nil
}

func (s *GarbageCollectorService) filterSnapshots(snapshotCleanupInfos []models.SnapshotCleanupInfo) []models.SnapshotCleanupInfo {
	var filtered []models.SnapshotCleanupInfo

	// Default cutoff time is 12 hours
	cutoffTime := time.Now().Add(-12 * time.Hour)
	maxAge, err := time.ParseDuration(s.excludeAge)
	if err == nil {
		cutoffTime = time.Now().Add(-maxAge)
	}

	for _, snapshotCleanupInfo := range snapshotCleanupInfos {
		if len(s.excludeSnapshots) > 0 {
			if slices.Contains(s.excludeSnapshots, strings.TrimSpace(snapshotCleanupInfo.Name)) {
				continue
			}
		}

		if !snapshotCleanupInfo.CreatedAt.Before(cutoffTime) {
			log.Printf("Excluding snapshot %s created at %s (within %v)", snapshotCleanupInfo.Name, snapshotCleanupInfo.CreatedAt.Format(time.RFC3339), maxAge)
			continue
		}

		filtered = append(filtered, snapshotCleanupInfo)
	}

	return filtered
}

func (s *GarbageCollectorService) checkSnapshotInAPI(ctx context.Context, snapshotName string) (bool, error) {
	runnersBySnapshotRefResponse, resp, err := s.client.RunnersAPI.GetRunnersBySnapshotRef(ctx).Ref(snapshotName).Execute()
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		// If 404, snapshot doesn't exist in API
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		log.Printf("Unexpected status code %d for snapshot %s", resp.StatusCode, snapshotName)

		return false, fmt.Errorf("unexpected status code %d for snapshot %s", resp.StatusCode, snapshotName)
	}

	// Check if any runner snapshot belongs to the expected domain
	for _, snapshot := range runnersBySnapshotRefResponse {
		if snapshot.RunnerDomain == s.domain {
			return true, nil
		}
	}

	// Snapshot exists but doesn't belong to expected domain
	if len(runnersBySnapshotRefResponse) > 0 {
		log.Printf("Snapshot %s exists but doesn't belong to expected runner domain %s", snapshotName, s.domain)
	}

	return false, nil
}
