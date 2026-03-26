package multiplexer

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner/pkg/volume"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC client for the volume multiplexer
type Client struct {
	conn *grpc.ClientConn
	// client api.VolumeMultiplexerClient // Will be enabled when proto is generated
}

// NewClient creates a new multiplexer client
func NewClient(address string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to multiplexer: %w", err)
	}

	return &Client{
		conn: conn,
		// client: api.NewVolumeMultiplexerClient(conn),
	}, nil
}

// RegisterVolume registers a volume with the multiplexer
func (c *Client) RegisterVolume(ctx context.Context, volumeID string, config volume.ProviderConfig, readOnly bool) error {
	// Convert to proto format
	protoConfig := &ProviderConfig{
		Type:       config.Type,
		Endpoint:   config.Endpoint,
		AccessKey:  config.AccessKey,
		SecretKey:  config.SecretKey,
		Region:     config.Region,
		BucketName: config.BucketName,
		Subpath:    config.Subpath,
		Options:    config.Options,
	}

	req := &RegisterVolumeRequest{
		VolumeId: volumeID,
		Config:   protoConfig,
		ReadOnly: readOnly,
	}

	// TODO: Call gRPC method when proto is generated
	// resp, err := c.client.RegisterVolume(ctx, req)
	// if err != nil {
	//     return err
	// }
	// if !resp.Success {
	//     return fmt.Errorf("registration failed: %s", resp.Error)
	// }

	_ = req // Temporary to avoid unused variable
	return nil
}

// UnregisterVolume removes a volume from the multiplexer
func (c *Client) UnregisterVolume(ctx context.Context, volumeID string) error {
	req := &UnregisterVolumeRequest{
		VolumeId: volumeID,
		Force:    false,
	}

	// TODO: Call gRPC method when proto is generated
	// _, err := c.client.UnregisterVolume(ctx, req)
	// return err

	_ = req
	return nil
}

// IncrementRefCount increments the reference count for a volume
func (c *Client) IncrementRefCount(ctx context.Context, volumeID string) error {
	req := &RefCountRequest{
		VolumeId: volumeID,
	}

	// TODO: Call gRPC method when proto is generated
	// _, err := c.client.IncrementRefCount(ctx, req)
	// return err

	_ = req
	return nil
}

// DecrementRefCount decrements the reference count for a volume
func (c *Client) DecrementRefCount(ctx context.Context, volumeID string) error {
	req := &RefCountRequest{
		VolumeId: volumeID,
	}

	// TODO: Call gRPC method when proto is generated
	// _, err := c.client.DecrementRefCount(ctx, req)
	// return err

	_ = req
	return nil
}

// GetStats retrieves multiplexer statistics
func (c *Client) GetStats(ctx context.Context) (*DaemonStats, error) {
	// TODO: Call gRPC method when proto is generated
	// resp, err := c.client.GetStats(ctx, &Empty{})
	// if err != nil {
	//     return nil, err
	// }

	// Convert proto stats to internal format
	return &DaemonStats{
		StartTime: time.Now(),
		Uptime:    0,
		// ... convert other fields
	}, nil
}

// HealthCheck checks if the multiplexer is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	// TODO: Call gRPC method when proto is generated
	// resp, err := c.client.HealthCheck(ctx, &Empty{})
	// if err != nil {
	//     return err
	// }
	// if !resp.Healthy {
	//     return fmt.Errorf("multiplexer unhealthy: %s", resp.Status)
	// }

	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
