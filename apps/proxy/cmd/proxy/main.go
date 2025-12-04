// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daytonaio/proxy/cmd/proxy/config"
	"github.com/daytonaio/proxy/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- proxy.StartProxy(ctx, config)
	}()

	var lastSignalTime time.Time

	for {
		select {
		case <-sigChan:
			if lastSignalTime.IsZero() {
				log.Info("Received shutdown, initiating graceful shutdown (press Ctrl+C again to force)")
				cancel()
				lastSignalTime = time.Now()
			} else if time.Since(lastSignalTime) < time.Millisecond {
				// If started as a subprocess, the app might receive multiple signals in quick succession instead of one
				// Debounce very closely spaced signals
				log.Info("Received second signal, but within debounce period, ignoring")
			} else {
				log.Info("Received second signal, forcing exit")
				os.Exit(1)
			}
		case err := <-errChan:
			if err != nil {
				log.Fatalf("Proxy exited with error: %v", err)
			} else {
				log.Info("Proxy exited gracefully")
				return
			}
		}
	}
}
