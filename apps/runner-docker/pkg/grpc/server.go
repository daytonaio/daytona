// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package grpc

import (
	"context"
	"log/slog"
	"net"
	"os"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/cache"
	"github.com/daytonaio/runner-docker/pkg/grpc/services/health"
	"github.com/daytonaio/runner-docker/pkg/grpc/services/sandbox"
	"github.com/daytonaio/runner-docker/pkg/grpc/services/snapshot"
	"github.com/docker/docker/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	loggerOpts = []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
)

type Server struct {
	grpcServer *grpc.Server
	address    string
}

type ServerConfig struct {
	Address            string
	DockerClient       *client.Client
	RunnerCache        *cache.IRunnerCache
	DaemonPath         string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSEndpointUrl     string
	Log                *slog.Logger
}

func New(cfg ServerConfig) *Server {
	log := cfg.Log.With("service", "grpc")

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			auth.UnaryServerInterceptor(authFn),
			logging.UnaryServerInterceptor(interceptorLogger(log), loggerOpts...),
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			auth.StreamServerInterceptor(authFn),
			logging.StreamServerInterceptor(interceptorLogger(log), loggerOpts...),
			recovery.StreamServerInterceptor(),
		),
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Initialize services
	healthSvc := health.NewHealthService(log)

	snapshotSvc := snapshot.NewSnapshotService(snapshot.SnapshotServiceConfig{
		DockerClient: cfg.DockerClient,
		Cache:        *cfg.RunnerCache,
		LogWriter:    os.Stdout,
		Log:          log,
	})

	sandboxSvc := sandbox.NewSandboxService(sandbox.SandboxServiceConfig{
		DockerClient:       cfg.DockerClient,
		SnapshotService:    snapshotSvc,
		Cache:              *cfg.RunnerCache,
		LogWriter:          os.Stdout,
		DaemonPath:         cfg.DaemonPath,
		AWSAccessKeyId:     cfg.AWSAccessKeyId,
		AWSSecretAccessKey: cfg.AWSSecretAccessKey,
		AWSRegion:          cfg.AWSRegion,
		AWSEndpointUrl:     cfg.AWSEndpointUrl,
		Log:                log,
	})

	// Register services
	pb.RegisterHealthServiceServer(grpcServer, healthSvc)
	pb.RegisterSandboxServiceServer(grpcServer, sandboxSvc)
	pb.RegisterSnapshotServiceServer(grpcServer, snapshotSvc)

	// Setup reflection
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		address:    cfg.Address,
	}
}

// Start server
func (s *Server) Start() error {
	// Create the TCP listener when starting the server
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	defer listener.Close()

	return s.grpcServer.Serve(listener)
}

// Shutdown server
func (s *Server) Shutdown(ctx context.Context) error {
	// Create a channel to signal when GracefulStop completes
	done := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for either GracefulStop to complete or context to be cancelled
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Context was cancelled, force stop the server
		s.grpcServer.Stop()
		return ctx.Err()
	}
}
