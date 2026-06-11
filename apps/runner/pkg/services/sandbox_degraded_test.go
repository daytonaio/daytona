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
