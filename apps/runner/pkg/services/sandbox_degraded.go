// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/docker"
)

const (
	// degradedProbeInterval is how often tracked sandboxes are re-probed.
	degradedProbeInterval = 30 * time.Second
	// degradedStaleAfter bounds how long a degraded flag survives without a
	// fresh confirmation (e.g. the container was recreated or the daemon
	// never recovers but stops reproducing the signature).
	degradedStaleAfter = 30 * time.Minute
	// degradedPushTimeout bounds a single push to the API.
	degradedPushTimeout = 10 * time.Second
)

type SandboxDegradedServiceConfig struct {
	Logger *slog.Logger
	Docker *docker.DockerClient
}

type degradedEntry struct {
	reason        string
	lastConfirmed time.Time
	reported      bool
	pushing       bool
}

// SandboxDegradedService tracks sandboxes observed in a degraded condition
// (currently file-descriptor exhaustion), pushes the degradedReason to the
// API, and runs a cheap daemon-exec probe over tracked sandboxes only,
// clearing the flag once the daemon can spawn processes again. Surfacing
// only — it never triggers recovery actions.
type SandboxDegradedService struct {
	log    *slog.Logger
	docker *docker.DockerClient

	mu      sync.Mutex
	entries map[string]*degradedEntry
	client  *apiclient.APIClient
}

func NewSandboxDegradedService(config SandboxDegradedServiceConfig) *SandboxDegradedService {
	return &SandboxDegradedService{
		log:     config.Logger.With(slog.String("component", "sandbox_degraded_service")),
		docker:  config.Docker,
		entries: make(map[string]*degradedEntry),
	}
}

func (s *SandboxDegradedService) getClient() (*apiclient.APIClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		client, err := runnerapiclient.GetApiClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get API client: %w", err)
		}
		s.client = client
	}
	return s.client, nil
}

// ReportFdExhaustion records an fd-exhaustion sighting for a sandbox. The
// reason must be the final degradedReason value (e.g. produced by
// common.ClassifyToolboxFdExhaustion). Cheap map upsert; the first sighting
// triggers an async push to the API (single-flight per sandbox).
func (s *SandboxDegradedService) ReportFdExhaustion(sandboxId string, reason string) {
	if sandboxId == "" || reason == "" {
		return
	}

	s.mu.Lock()
	entry, ok := s.entries[sandboxId]
	if !ok {
		entry = &degradedEntry{}
		s.entries[sandboxId] = entry
	}
	entry.reason = reason
	entry.lastConfirmed = time.Now()
	shouldPush := !entry.reported && !entry.pushing
	if shouldPush {
		entry.pushing = true
	}
	s.mu.Unlock()

	if !shouldPush {
		return
	}

	go func() {
		err := s.pushReason(sandboxId, reason)

		s.mu.Lock()
		if entry, ok := s.entries[sandboxId]; ok {
			entry.pushing = false
			entry.reported = err == nil
		}
		s.mu.Unlock()

		if err != nil {
			s.log.Warn("Failed to push degraded reason, will retry on next probe tick", "sandboxId", sandboxId, "error", err)
		} else {
			s.log.Info("Marked sandbox degraded", "sandboxId", sandboxId, "reason", reason)
		}
	}()
}

// Start seeds tracking from the API (healing runner restarts) and runs the
// probe loop until ctx is done.
func (s *SandboxDegradedService) Start(ctx context.Context) {
	s.log.InfoContext(ctx, "Starting sandbox degraded tracking")
	go func() {
		s.seedFromApi(ctx)

		ticker := time.NewTicker(degradedProbeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.probeTrackedSandboxes(ctx)
			case <-ctx.Done():
				s.log.InfoContext(ctx, "Sandbox degraded service stopped")
				return
			}
		}
	}()
}

// seedFromApi restores tracking for sandboxes that already carry a
// degradedReason, so a runner restart cannot strand a stale flag. Best-effort.
func (s *SandboxDegradedService) seedFromApi(ctx context.Context) {
	client, err := s.getClient()
	if err != nil {
		s.log.WarnContext(ctx, "Failed to seed degraded sandboxes", "error", err)
		return
	}

	sandboxes, _, err := client.SandboxAPI.GetSandboxesForRunner(ctx).
		States(string(apiclient.SANDBOXSTATE_STARTED)).
		Execute()
	if err != nil {
		s.log.WarnContext(ctx, "Failed to seed degraded sandboxes", "error", err)
		return
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sandbox := range sandboxes {
		if sandbox.Id == "" || sandbox.DegradedReason == nil || *sandbox.DegradedReason == "" {
			continue
		}
		if _, ok := s.entries[sandbox.Id]; ok {
			continue
		}
		s.entries[sandbox.Id] = &degradedEntry{
			reason:        *sandbox.DegradedReason,
			lastConfirmed: now,
			reported:      true,
		}
		s.log.InfoContext(ctx, "Restored degraded sandbox tracking", "sandboxId", sandbox.Id, "reason", *sandbox.DegradedReason)
	}
}

func (s *SandboxDegradedService) probeTrackedSandboxes(ctx context.Context) {
	s.mu.Lock()
	sandboxIds := make([]string, 0, len(s.entries))
	for sandboxId := range s.entries {
		sandboxIds = append(sandboxIds, sandboxId)
	}
	s.mu.Unlock()

	for _, sandboxId := range sandboxIds {
		s.probeSandbox(ctx, sandboxId)
	}
}

func (s *SandboxDegradedService) probeSandbox(ctx context.Context, sandboxId string) {
	// Drop tracking without pushing when the container is gone or not
	// running — the API-side invariant already clears the field on the state
	// transition.
	c, err := s.docker.ContainerInspect(ctx, sandboxId)
	if err != nil || c == nil || c.State == nil || !c.State.Running {
		s.drop(sandboxId)
		s.log.DebugContext(ctx, "Dropped degraded tracking for non-running sandbox", "sandboxId", sandboxId)
		return
	}

	healthy, observed, err := s.docker.ProbeDaemonExec(ctx, sandboxId)

	switch {
	case err == nil && healthy:
		// The daemon can spawn processes again — clear the flag. Skip the
		// clear while a reason push is in flight: the push could land after
		// the clear, stranding a degradedReason on a healthy sandbox with no
		// tracked entry left to heal it. The entry stays tracked and the
		// clear is retried next tick.
		if s.pushInFlight(sandboxId) {
			s.log.DebugContext(ctx, "Skipping degraded clear while reason push is in flight", "sandboxId", sandboxId)
			return
		}
		if pushErr := s.pushClear(sandboxId); pushErr != nil {
			s.log.WarnContext(ctx, "Failed to clear degraded reason, will retry on next probe tick", "sandboxId", sandboxId, "error", pushErr)
			return
		}
		s.drop(sandboxId)
		s.log.InfoContext(ctx, "Cleared degraded reason", "sandboxId", sandboxId)
	case err == nil && common.MatchFdExhaustion(observed):
		// Still degraded — refresh confirmation and make sure it is reported.
		s.refresh(sandboxId, common.FdExhaustionReason(observed))
	default:
		// Indeterminate (transport/decode failure) or a non-fd exec failure:
		// keep the entry, but force-clear once it has gone stale.
		s.expireIfStale(ctx, sandboxId)
	}
}

func (s *SandboxDegradedService) refresh(sandboxId string, reason string) {
	s.mu.Lock()
	entry, ok := s.entries[sandboxId]
	if !ok {
		s.mu.Unlock()
		return
	}
	entry.reason = reason
	entry.lastConfirmed = time.Now()
	needsPush := !entry.reported && !entry.pushing
	s.mu.Unlock()

	if !needsPush {
		return
	}

	if err := s.pushReason(sandboxId, reason); err != nil {
		s.log.Warn("Failed to push degraded reason, will retry on next probe tick", "sandboxId", sandboxId, "error", err)
		return
	}

	s.mu.Lock()
	if entry, ok := s.entries[sandboxId]; ok {
		entry.reported = true
	}
	s.mu.Unlock()
}

// pushInFlight reports whether an async degradedReason push is still running
// for the sandbox.
func (s *SandboxDegradedService) pushInFlight(sandboxId string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[sandboxId]
	return ok && entry.pushing
}

func (s *SandboxDegradedService) expireIfStale(ctx context.Context, sandboxId string) {
	s.mu.Lock()
	entry, ok := s.entries[sandboxId]
	stale := ok && time.Since(entry.lastConfirmed) > degradedStaleAfter
	s.mu.Unlock()

	if !stale {
		return
	}

	if err := s.pushClear(sandboxId); err != nil {
		s.log.WarnContext(ctx, "Failed to clear stale degraded reason, will retry on next probe tick", "sandboxId", sandboxId, "error", err)
		return
	}
	s.drop(sandboxId)
	s.log.InfoContext(ctx, "Cleared stale degraded reason", "sandboxId", sandboxId)
}

func (s *SandboxDegradedService) drop(sandboxId string) {
	s.mu.Lock()
	delete(s.entries, sandboxId)
	s.mu.Unlock()
}

func (s *SandboxDegradedService) pushReason(sandboxId string, reason string) error {
	dto := apiclient.NewUpdateSandboxDegradedReasonDto()
	dto.SetDegradedReason(reason)
	return s.push(sandboxId, *dto)
}

func (s *SandboxDegradedService) pushClear(sandboxId string) error {
	dto := apiclient.NewUpdateSandboxDegradedReasonDto()
	dto.SetDegradedReasonNil()
	return s.push(sandboxId, *dto)
}

func (s *SandboxDegradedService) push(sandboxId string, dto apiclient.UpdateSandboxDegradedReasonDto) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), degradedPushTimeout)
	defer cancel()

	_, err = client.SandboxAPI.UpdateSandboxDegradedReason(ctx, sandboxId).
		UpdateSandboxDegradedReasonDto(dto).
		Execute()
	return err
}
