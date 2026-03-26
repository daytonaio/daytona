package multiplexer

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/daytonaio/runner/pkg/volume"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the VolumeMultiplexer gRPC service
type GRPCServer struct {
	daemon     *MultiplexerDaemon
	grpcServer *grpc.Server
	logger     *slog.Logger
}

// NewGRPCServer creates a new gRPC server for the multiplexer
func NewGRPCServer(daemon *MultiplexerDaemon, logger *slog.Logger) *GRPCServer {
	return &GRPCServer{
		daemon: daemon,
		logger: logger,
	}
}

// Start starts the gRPC server on the specified address
func (s *GRPCServer) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = grpc.NewServer()
	// Register service here when proto is generated
	// api.RegisterVolumeMultiplexerServer(s.grpcServer, s)

	s.logger.Info("Starting gRPC server", "address", address)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// RegisterVolume adds a volume to the multiplexer
func (s *GRPCServer) RegisterVolume(ctx context.Context, req *RegisterVolumeRequest) (*RegisterVolumeResponse, error) {
	s.logger.Debug("RegisterVolume called", "volumeID", req.VolumeId)

	// Convert proto config to internal config
	config := volume.ProviderConfig{
		Type:       req.Config.Type,
		Endpoint:   req.Config.Endpoint,
		AccessKey:  req.Config.AccessKey,
		SecretKey:  req.Config.SecretKey,
		Region:     req.Config.Region,
		BucketName: req.Config.BucketName,
		Subpath:    req.Config.Subpath,
		Options:    req.Config.Options,
	}

	// Register with daemon
	err := s.daemon.RegisterVolume(ctx, req.VolumeId, config, req.ReadOnly)
	if err != nil {
		s.logger.Error("Failed to register volume", "volumeID", req.VolumeId, "error", err)
		return &RegisterVolumeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &RegisterVolumeResponse{
		Success: true,
	}, nil
}

// UnregisterVolume removes a volume from the multiplexer
func (s *GRPCServer) UnregisterVolume(ctx context.Context, req *UnregisterVolumeRequest) (*Empty, error) {
	s.logger.Debug("UnregisterVolume called", "volumeID", req.VolumeId)

	err := s.daemon.UnregisterVolume(ctx, req.VolumeId)
	if err != nil {
		if req.Force {
			s.logger.Warn("Force unregistering volume", "volumeID", req.VolumeId, "error", err)
			// Implement force removal logic
		} else {
			return nil, status.Errorf(codes.FailedPrecondition, "failed to unregister volume: %v", err)
		}
	}

	return &Empty{}, nil
}

// IncrementRefCount increases the reference count for a volume
func (s *GRPCServer) IncrementRefCount(ctx context.Context, req *RefCountRequest) (*Empty, error) {
	err := s.daemon.IncrementRefCount(req.VolumeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "volume not found: %s", req.VolumeId)
	}
	return &Empty{}, nil
}

// DecrementRefCount decreases the reference count for a volume
func (s *GRPCServer) DecrementRefCount(ctx context.Context, req *RefCountRequest) (*Empty, error) {
	err := s.daemon.DecrementRefCount(req.VolumeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "volume not found: %s", req.VolumeId)
	}
	return &Empty{}, nil
}

// GetStats returns current daemon statistics
func (s *GRPCServer) GetStats(ctx context.Context, req *Empty) (*MultiplexerStats, error) {
	stats := s.daemon.GetStats()

	// Convert internal stats to proto format
	protoStats := &MultiplexerStats{
		StartTime:         stats.StartTime.Unix(),
		UptimeSeconds:     int64(stats.Uptime.Seconds()),
		TotalVolumes:      int32(stats.TotalVolumes),
		ActiveVolumes:     int32(stats.ActiveVolumes),
		TotalReads:        stats.TotalReads,
		TotalWrites:       stats.TotalWrites,
		TotalBytesRead:    stats.TotalBytesRead,
		TotalBytesWritten: stats.TotalBytesWrite,
		CacheHitRate:      stats.CacheHitRate,
	}

	// Convert volume stats
	for _, vs := range stats.VolumeStats {
		protoStats.VolumeStats = append(protoStats.VolumeStats, &VolumeStats{
			VolumeId:          vs.VolumeID,
			RegisteredAt:      vs.RegisteredAt.Unix(),
			LastAccessedAt:    vs.LastAccessedAt.Unix(),
			ReadOperations:    vs.ReadOperations,
			WriteOperations:   vs.WriteOperations,
			BytesRead:         vs.BytesRead,
			BytesWritten:      vs.BytesWritten,
			CacheHits:         vs.CacheHits,
			CacheMisses:       vs.CacheMisses,
			ActiveFileHandles: vs.ActiveFileHandles,
		})
	}

	return protoStats, nil
}

// HealthCheck returns the health status of the daemon
func (s *GRPCServer) HealthCheck(ctx context.Context, req *Empty) (*HealthStatus, error) {
	// TODO: Implement actual health checks
	return &HealthStatus{
		Healthy:   true,
		Status:    "OK",
		LastCheck: time.Now().Unix(),
	}, nil
}

// Temporary type definitions until proto is generated
type RegisterVolumeRequest struct {
	VolumeId string
	Config   *ProviderConfig
	ReadOnly bool
}

type RegisterVolumeResponse struct {
	Success bool
	Error   string
}

type UnregisterVolumeRequest struct {
	VolumeId string
	Force    bool
}

type RefCountRequest struct {
	VolumeId string
}

type ProviderConfig struct {
	Type       string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Region     string
	BucketName string
	Subpath    string
	Options    map[string]string
}

type MultiplexerStats struct {
	StartTime         int64
	UptimeSeconds     int64
	TotalVolumes      int32
	ActiveVolumes     int32
	TotalReads        uint64
	TotalWrites       uint64
	TotalBytesRead    uint64
	TotalBytesWritten uint64
	CacheHitRate      float64
	VolumeStats       []*VolumeStats
}

type VolumeStats struct {
	VolumeId          string
	RegisteredAt      int64
	LastAccessedAt    int64
	ReadOperations    uint64
	WriteOperations   uint64
	BytesRead         uint64
	BytesWritten      uint64
	CacheHits         uint64
	CacheMisses       uint64
	ActiveFileHandles int32
}

type HealthStatus struct {
	Healthy   bool
	Status    string
	LastCheck int64
}

type Empty struct{}
