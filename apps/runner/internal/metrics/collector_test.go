/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/docker/docker/api/types/system"
)

func TestDockerInfoAndDataRootUsesConfiguredOverride(t *testing.T) {
	ctx := context.Background()
	infoCalls := 0
	collector := newTestCollector(func(context.Context) (system.Info, error) {
		infoCalls++
		return system.Info{
			DockerRootDir: "/data/docker",
			Images:        17,
		}, nil
	})
	collector.dockerDataRoot = "/mnt/runner/docker"

	info, dataRoot, err := collector.dockerInfoAndDataRoot(ctx)
	if err != nil {
		t.Fatalf("dockerInfoAndDataRoot returned error: %v", err)
	}
	if infoCalls != 1 {
		t.Fatalf("expected one Docker info call for snapshot count, got %d", infoCalls)
	}
	if dataRoot != "/mnt/runner/docker" {
		t.Fatalf("expected configured data root, got %q", dataRoot)
	}
	if info.Images != 17 {
		t.Fatalf("expected Docker info to be returned, got Images=%d", info.Images)
	}
}

func TestDockerInfoAndDataRootUsesDockerReportedRoot(t *testing.T) {
	ctx := context.Background()
	infoCalls := 0
	collector := newTestCollector(func(context.Context) (system.Info, error) {
		infoCalls++
		return system.Info{
			DockerRootDir: "/data/docker",
			Images:        3,
		}, nil
	})

	info, dataRoot, err := collector.dockerInfoAndDataRoot(ctx)
	if err != nil {
		t.Fatalf("dockerInfoAndDataRoot returned error: %v", err)
	}
	if infoCalls != 1 {
		t.Fatalf("expected one Docker info call, got %d", infoCalls)
	}
	if dataRoot != "/data/docker" {
		t.Fatalf("expected Docker-reported data root, got %q", dataRoot)
	}
	if info.Images != 3 {
		t.Fatalf("expected Docker info to be returned, got Images=%d", info.Images)
	}
}

func TestDockerInfoAndDataRootFallsBackWhenInfoFails(t *testing.T) {
	ctx := context.Background()
	infoErr := errors.New("docker unavailable")
	collector := newTestCollector(func(context.Context) (system.Info, error) {
		return system.Info{}, infoErr
	})

	_, dataRoot, err := collector.dockerInfoAndDataRoot(ctx)
	if !errors.Is(err, infoErr) {
		t.Fatalf("expected Docker info error to be returned, got %v", err)
	}
	if dataRoot != defaultDockerDataRoot {
		t.Fatalf("expected fallback data root %q, got %q", defaultDockerDataRoot, dataRoot)
	}
}

func newTestCollector(dockerInfo func(context.Context) (system.Info, error)) *Collector {
	return &Collector{
		log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		dockerInfo: dockerInfo,
	}
}
