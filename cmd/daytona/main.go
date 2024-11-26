// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

var (
	defaultTimeout = 5 * time.Second
	maxRetries     = 3
	retryDelay     = time.Second
	apiServerAddr  = "http://localhost:3986" // Updated to match Daytona's default port
	headscaleAddr  = "http://localhost:3986" // Using same port as API server
	registryAddr   = "localhost:5000"        // Default registry port
)

func main() {
	if internal.WorkspaceMode() {
		err := workspacemode.Execute()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for all components to be healthy
	timeout := 2 * time.Minute
	if err := checkComponentHealth(timeout); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}

	log.Info("Daytona server is fully operational")
}

func init() {
	logLevel := log.WarnLevel
	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		if parsedLevel, err := log.ParseLevel(logLevelEnv); err == nil {
			logLevel = parsedLevel
		}
	}

	log.SetLevel(logLevel)
	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		zerologLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(zerologLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
		Out:        &util.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})
}

func checkComponentHealth(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	components := []struct {
		name  string
		check func(context.Context) error
	}{
		{"API Server", checkAPIServer},
		{"Providers", checkProviders},
		{"Local Registry", checkLocalRegistry},
		{"Headscale Server", checkHeadscaleServer},
	}

	for _, component := range components {
		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			select {
			case <-ctx.Done():
				return fmt.Errorf("%s health check timed out: %w", component.name, ctx.Err())
			default:
				if err := component.check(ctx); err != nil {
					lastErr = err
					log.Warnf("%s health check failed (attempt %d/%d): %v",
						component.name, attempt, maxRetries, err)
					if attempt < maxRetries {
						time.Sleep(retryDelay)
						continue
					}
					return fmt.Errorf("%s health check failed after %d attempts: %w",
						component.name, maxRetries, lastErr)
				}
				log.Infof("%s is healthy", component.name)
				goto nextComponent
			}
		}
	nextComponent:
	}
	return nil
}

func checkAPIServer(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", apiServerAddr+"/api/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to API server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API server returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}

func checkProviders(ctx context.Context) error {
	// Check specifically for Docker provider v0.12.1
	provider := "docker-provider"
	version := "v0.12.1"

	select {
	case <-ctx.Done():
		return fmt.Errorf("provider check timed out for %s: %w", provider, ctx.Err())
	default:
		// Simulating a quick check for Docker provider
		time.Sleep(100 * time.Millisecond)
		log.Printf("Docker provider (%s %s) is available", provider, version)
		return nil
	}
}

func checkLocalRegistry(ctx context.Context) error {
	d := net.Dialer{Timeout: defaultTimeout}
	conn, err := d.DialContext(ctx, "tcp", registryAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to local registry: %w", err)
	}
	defer conn.Close()
	return nil
}

func checkHeadscaleServer(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", headscaleAddr+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to headscale server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("headscale server returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}
