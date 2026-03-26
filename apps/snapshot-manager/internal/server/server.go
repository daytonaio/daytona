/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/daytonaio/snapshot-manager/internal"
	"github.com/distribution/distribution/v3/configuration"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	"github.com/distribution/distribution/v3/registry/handlers"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
)

type Server struct {
	app    *handlers.App
	config *configuration.Configuration
	server *http.Server
	logger *slog.Logger
}

// NewServer creates a new registry server with the given configuration
func NewServer(config *configuration.Configuration, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Validate storage configuration
	if len(config.Storage) == 0 {
		return nil, fmt.Errorf("storage configuration is required")
	}

	// Create storage directory if using filesystem driver
	if fsParams, ok := config.Storage["filesystem"]; ok {
		if rootDir, ok := fsParams["rootdirectory"].(string); ok {
			if err := os.MkdirAll(rootDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create storage directory: %w", err)
			}
		}
	}

	return &Server{
		config: config,
		logger: logger,
	}, nil
}

// Start starts the registry server and blocks until context is cancelled
func (s *Server) Start(ctx context.Context) error {
	// Create the registry app
	app := handlers.NewApp(ctx, s.config)
	s.app = app

	// Create HTTP server with health check endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.Handle("/", app)

	s.server = &http.Server{
		Addr:    s.config.HTTP.Addr,
		Handler: mux,
	}

	s.logger.Info("Daytona snapshot-manager started",
		slog.String("version", internal.Version),
		slog.String("addr", s.config.HTTP.Addr),
		slog.Any("storage", s.getStorageInfo()),
	)

	// Start the server
	errChan := make(chan error, 1)
	go func() {
		var err error
		if s.config.HTTP.TLS.Certificate != "" {
			s.logger.Info("Starting with TLS",
				slog.String("cert", s.config.HTTP.TLS.Certificate),
			)
			err = s.server.ListenAndServeTLS(s.config.HTTP.TLS.Certificate, s.config.HTTP.TLS.Key)
		} else {
			err = s.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

// Shutdown gracefully shuts down the registry server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Daytona snapshot-manager...")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// getStorageInfo returns a human-readable storage configuration info
func (s *Server) getStorageInfo() map[string]interface{} {
	info := make(map[string]interface{})
	for driver := range s.config.Storage {
		if driver == "delete" || driver == "maintenance" || driver == "cache" {
			continue
		}
		info["driver"] = driver
		if params, ok := s.config.Storage[driver]; ok {
			// Add relevant params based on driver
			switch driver {
			case "filesystem":
				if rootDir, ok := params["rootdirectory"].(string); ok {
					info["rootdirectory"] = rootDir
				}
			case "s3":
				if region, ok := params["region"].(string); ok {
					info["region"] = region
				}
				if bucket, ok := params["bucket"].(string); ok {
					info["bucket"] = bucket
				}
			}
		}
	}
	return info
}
