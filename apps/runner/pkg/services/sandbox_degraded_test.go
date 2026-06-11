// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/docker/docker/api/types/container"
)

func TestPushInFlight(t *testing.T) {
	cases := []struct {
		name    string
		entries map[string]*degradedEntry
		want    bool
	}{
		{
			name:    "untracked sandbox is not in flight",
			entries: map[string]*degradedEntry{},
			want:    false,
		},
		{
			name:    "tracked entry without active push is not in flight",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", pushing: false}},
			want:    false,
		},
		{
			name:    "tracked entry with active push blocks the clear",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", pushing: true}},
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &SandboxDegradedService{entries: tc.entries}
			if got := s.pushInFlight("sb"); got != tc.want {
				t.Errorf("pushInFlight() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClaimReasonPush(t *testing.T) {
	cases := []struct {
		name    string
		entries map[string]*degradedEntry
		want    bool
	}{
		{
			name:    "untracked sandbox is not claimable",
			entries: map[string]*degradedEntry{},
			want:    false,
		},
		{
			name:    "unreported entry is claimed",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion"}},
			want:    true,
		},
		{
			name:    "reported entry needs no push",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", reported: true}},
			want:    false,
		},
		{
			name:    "in-flight push is not claimed twice",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", pushing: true}},
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &SandboxDegradedService{entries: tc.entries}
			if got := s.claimReasonPush("sb"); got != tc.want {
				t.Fatalf("claimReasonPush() = %v, want %v", got, tc.want)
			}
			if tc.want {
				if entry := tc.entries["sb"]; !entry.pushing {
					t.Error("claimReasonPush() must set entry.pushing so clear paths can see the in-flight push")
				}
				if s.claimReasonPush("sb") {
					t.Error("second claimReasonPush() must be refused while the push is in flight")
				}
			}
		})
	}
}

// newDegradedTestService builds a service whose API client points at the
// given test server, so push paths are exercised over real HTTP.
func newDegradedTestService(t *testing.T, serverUrl string) *SandboxDegradedService {
	t.Helper()
	cfg := apiclient.NewConfiguration()
	cfg.Servers = apiclient.ServerConfigurations{{URL: serverUrl}}
	return &SandboxDegradedService{
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		entries: make(map[string]*degradedEntry),
		client:  apiclient.NewAPIClient(cfg),
	}
}

// TestRefreshHoldsPushGuard pins the refresh re-push to the same pushing
// guard the report path uses: while the push is on the wire, pushInFlight
// must be true (so clear paths skip), and completion must mark the entry
// reported and release the guard.
func TestRefreshHoldsPushGuard(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newDegradedTestService(t, srv.URL)
	s.entries["sb"] = &degradedEntry{reason: "fd exhaustion", lastConfirmed: time.Now()}

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.refresh("sb", "fd exhaustion")
	}()

	<-started
	if !s.pushInFlight("sb") {
		t.Error("refresh must hold the pushing guard while its reason push is in flight")
	}
	close(release)
	<-done

	s.mu.Lock()
	entry := s.entries["sb"]
	s.mu.Unlock()
	if entry == nil || !entry.reported || entry.pushing {
		t.Fatalf("refresh must mark the entry reported and release the guard, got %+v", entry)
	}
}

func TestExpireIfStale(t *testing.T) {
	stale := time.Now().Add(-degradedStaleAfter - time.Minute)
	cases := []struct {
		name      string
		entry     *degradedEntry
		wantCalls int32
		wantKept  bool
	}{
		{
			name:      "fresh entry is not expired",
			entry:     &degradedEntry{reason: "fd exhaustion", lastConfirmed: time.Now()},
			wantCalls: 0,
			wantKept:  true,
		},
		{
			name:      "stale entry with push in flight is kept for the next tick",
			entry:     &degradedEntry{reason: "fd exhaustion", lastConfirmed: stale, pushing: true},
			wantCalls: 0,
			wantKept:  true,
		},
		{
			name:      "stale entry is cleared and dropped",
			entry:     &degradedEntry{reason: "fd exhaustion", lastConfirmed: stale},
			wantCalls: 1,
			wantKept:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			s := newDegradedTestService(t, srv.URL)
			s.entries["sb"] = tc.entry

			s.expireIfStale(context.Background(), "sb")

			if got := calls.Load(); got != tc.wantCalls {
				t.Errorf("expireIfStale() pushed %d clears, want %d", got, tc.wantCalls)
			}
			if _, ok := s.entries["sb"]; ok != tc.wantKept {
				t.Errorf("expireIfStale() kept entry = %v, want %v", ok, tc.wantKept)
			}
		})
	}
}

func TestClassifyInspect(t *testing.T) {
	cases := []struct {
		name string
		c    *container.InspectResponse
		err  error
		want probeAction
	}{
		{
			name: "container not found drops tracking",
			err:  common_errors.NewNotFoundError(errors.New("failed to inspect sandbox container sb")),
			want: probeActionDrop,
		},
		{
			name: "transient inspect failure keeps the entry",
			err:  errors.New("failed to inspect sandbox container sb: cannot connect to the Docker daemon"),
			want: probeActionRetry,
		},
		{
			name: "nil response drops tracking",
			c:    nil,
			want: probeActionDrop,
		},
		{
			name: "missing state drops tracking",
			c:    &container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{}},
			want: probeActionDrop,
		},
		{
			name: "stopped container drops tracking",
			c:    &container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{State: &container.State{Running: false}}},
			want: probeActionDrop,
		},
		{
			name: "running container is probed",
			c:    &container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{State: &container.State{Running: true}}},
			want: probeActionProbe,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyInspect(tc.c, tc.err); got != tc.want {
				t.Errorf("classifyInspect() = %v, want %v", got, tc.want)
			}
		})
	}
}

// waitForGuardRelease polls until the entry's pushing guard is released and
// returns a snapshot of the entry.
func waitForGuardRelease(t *testing.T, s *SandboxDegradedService, sandboxId string) degradedEntry {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		entry, ok := s.entries[sandboxId]
		var snapshot degradedEntry
		if ok {
			snapshot = *entry
		}
		s.mu.Unlock()
		if ok && !snapshot.pushing {
			return snapshot
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("pushing guard for %s was not released within deadline", sandboxId)
	return degradedEntry{}
}

// TestReportFdExhaustionClaimsBeforeSpawn pins the spawn-then-claim ordering:
// the pushing guard must be observable immediately after ReportFdExhaustion
// returns, before the push goroutine has run. If the goroutine claimed
// instead, a probe tick in the spawn window would observe pushing=false,
// push a clear, and the late reason push would strand the flag.
func TestReportFdExhaustionClaimsBeforeSpawn(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newDegradedTestService(t, srv.URL)

	s.ReportFdExhaustion("sb", "fd-exhaustion: too many open files")

	// Synchronous check, deliberately before the push goroutine is allowed
	// to make progress (the server blocks until release is closed).
	if !s.pushInFlight("sb") {
		t.Fatal("pushing guard must be claimed before ReportFdExhaustion returns")
	}

	<-started
	close(release)

	entry := waitForGuardRelease(t, s, "sb")
	if !entry.reported {
		t.Fatalf("push completion must mark the entry reported, got %+v", entry)
	}
}

// TestReportFdExhaustionRetriesAfterRejectedPush pins the push error path: a
// non-2xx response (e.g. the API's 409 while the sandbox is not STARTED yet)
// must leave the entry unreported with the guard released, so the next probe
// tick's refresh retries the push — and the retry succeeds once the API
// accepts it.
func TestReportFdExhaustionRetriesAfterRejectedPush(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newDegradedTestService(t, srv.URL)
	s.ReportFdExhaustion("sb", "fd-exhaustion: too many open files")

	entry := waitForGuardRelease(t, s, "sb")
	if entry.reported {
		t.Fatal("a rejected push must not mark the entry reported")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("first push: got %d calls, want 1", got)
	}

	// The next probe tick re-confirms the condition and refreshes — the
	// unreported entry must be re-pushed.
	s.refresh("sb", "fd-exhaustion: too many open files")

	if got := calls.Load(); got != 2 {
		t.Fatalf("retry push: got %d calls, want 2", got)
	}
	s.mu.Lock()
	reported := s.entries["sb"] != nil && s.entries["sb"].reported
	s.mu.Unlock()
	if !reported {
		t.Fatal("successful retry must mark the entry reported")
	}
}

// TestSeedFromApiReturnsErrorForRetry pins the contract the startup retry
// loop depends on: a failed seed must surface an error instead of being
// swallowed, so seedFromApiWithRetry can re-attempt it.
func TestSeedFromApiReturnsErrorForRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := newDegradedTestService(t, srv.URL)
	if err := s.seedFromApi(context.Background()); err == nil {
		t.Fatal("seedFromApi must return an error so the startup retry loop can re-attempt")
	}
}

// TestSeedFromApiWithRetrySucceedsAfterFailures pins the startup retry loop:
// it keeps re-attempting through transient API failures (there is no give-up
// path) and returns once a seed succeeds.
func TestSeedFromApiWithRetrySucceedsAfterFailures(t *testing.T) {
	prevBackoff := degradedSeedInitialBackoff
	degradedSeedInitialBackoff = time.Millisecond
	t.Cleanup(func() { degradedSeedInitialBackoff = prevBackoff })

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	s := newDegradedTestService(t, srv.URL)
	done := make(chan struct{})
	go func() {
		s.seedFromApiWithRetry(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("seedFromApiWithRetry did not return after a successful seed")
	}
	if got := calls.Load(); got != 4 {
		t.Fatalf("got %d seed attempts, want 4 (3 failures, then success)", got)
	}
}
